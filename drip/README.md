# Drip (server)

Drip enables server-client communication with custom state serialization
and transmission. The server runs on a fixed timestep and handles core game logic,
while clients receive state updates and can send input to the server. The client may also run the
main sim when network buffer is stale.

Currently in early development. Only supports networking for a single scene. Uses a basic server authoritative
TCP architecture.

## Basic usage

### Server Example

```go

package main

import (
 "log"
 "os"
 "os/signal"
 "syscall"

 "github.com/TheBitDrifter/bappa/drip"
 "github.com/TheBitDrifter/bappa/drip/drip_seversystems"
 "example/shared/coresystems"
 "example/shared/scenes"
)

func main() {
 drip.Callbacks.NewConnectionCreateEntity = NewConnectionEntityCreate
 drip.Callbacks.Serialize = SerializeCallback

 config := drip.DefaultServerConfig()

 server := drip.NewServer(config, drip_seversystems.ActionBufferSystem{})

 // Register a scene
 log.Println("Registering scene:", scenes.SceneOne.Name)
 err := server.RegisterScene(
  scenes.SceneOne.Name,
  scenes.SceneOne.Width,
  scenes.SceneOne.Height,
  scenes.SceneOne.Plan,
  coresystems.DefaultCoreSystems,
 )
 if err != nil {
  log.Fatalf("Failed to register scene: %v", err)
 }

 // Start the server
 log.Println("Starting server...")
 if err := server.Start(); err != nil {
  log.Fatalf("Failed to start server: %v", err)
 }

 // Create a channel to receive OS signals
 quit := make(chan os.Signal, 1)

 // Notify the channel for specific signals (Interrupt, Terminate)
 signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 log.Println("Server running. Press Ctrl+C to stop.")

 // Block execution until a signal is received on the 'quit' channel
 <-quit

 // Initiate shutdown
 log.Println("Shutting down server...")
 if err := server.Stop(); err != nil {
  log.Printf("Error stopping server: %v", err)
 } else {
  log.Println("Server stopped gracefully.")
 }
}
```

### Client Example

```go
package main

import (
 "log"

 "github.com/TheBitDrifter/bappa/blueprint"
 "github.com/TheBitDrifter/bappa/coldbrew"
 "github.com/TheBitDrifter/bappa/coldbrew/coldbrew_clientsystems"
 "github.com/TheBitDrifter/bappa/coldbrew/coldbrew_rendersystems"

 "example/shared/actions"
 "example/shared/scenes"
 "example/sharedclient"
 "example/sharedclient/assets"
 "example/sharedclient/clientsystems"
 "example/sharedclient/rendersystems"
 "example/shared/coresystems" // optional (if you want client to interpolate/run core sim when network buffer is stale)

 "github.com/hajimehoshi/ebiten/v2"
)

func main() {
 log.Println("Starting Networked Client...")

 client := coldbrew.NewNetworkClient(
  sharedclient.RESOLUTION_X,
  sharedclient.RESOLUTION_Y,
  sharedclient.MAX_SPRITES_CACHED,
  sharedclient.MAX_SOUNDS_CACHED,
  sharedclient.MAX_SCENES_CACHED,
  assets.FS,
 )

 client.SetDeserCallback(Derser)

 client.SetLocalAssetPath("../sharedclient/assets/")

 // Client Settings
 client.SetTitle("Platformer LDTK Template (Networked)")
 client.SetResizable(true)
 client.SetMinimumLoadTime(30)

 log.Println("Registering Scene One...")
 err := client.RegisterScene(
  scenes.SceneOne.Name,
  scenes.SceneOne.Width,
  scenes.SceneOne.Height,
  scenes.SceneOne.Plan,
  rendersystems.DefaultRenderSystems,
  clientsystems.DefaultClientSystemsNetworked,
  coresystems.DefaultCoreSystems{}, // OR
  // []blueprint.CoreSystem{}, <- use this instead for no interpolation
  scenes.SceneOne.Preload...,
 )
 if err != nil {
  log.Fatalf("Failed to register Scene One: %v", err)
 }

 // Register Global Systems
 log.Println("Registering Global Systems...")
 client.RegisterGlobalRenderSystem(
  coldbrew_rendersystems.GlobalRenderer{},
  &coldbrew_rendersystems.DebugRenderer{},
 )
 client.RegisterGlobalClientSystem(
  &coldbrew_clientsystems.InputSenderSystem{},
  coldbrew_clientsystems.InputBufferSystem{},
  &coldbrew_clientsystems.CameraSceneAssignerSystem{},
 )

 log.Println("Activating Camera...")
 _, err = client.ActivateCamera()
 if err != nil {
  log.Fatalf("Failed to activate camera: %v", err)
 }

 log.Println("Activating Input Receiver and Mapping Keys...")
 receiver1, err := client.ActivateReceiver()
 if err != nil {
  log.Fatalf("Failed to activate receiver: %v", err)
 }
 receiver1.RegisterKey(ebiten.KeySpace, actions.Jump)
 receiver1.RegisterKey(ebiten.KeyW, actions.Jump)
 receiver1.RegisterKey(ebiten.KeyA, actions.Left)
 receiver1.RegisterKey(ebiten.KeyD, actions.Right)
 receiver1.RegisterKey(ebiten.KeyS, actions.Down)

 log.Printf("Connecting to Drip server at %s...", sharedclient.SERVER_ADDRESS)
 err = client.Connect(sharedclient.SERVER_ADDRESS)
 if err != nil {
  log.Fatalf("Failed to connect to server '%s': %v", sharedclient.SERVER_ADDRESS, err)
 }
 defer func() {
  log.Println("Disconnecting from server...")
  client.Disconnect()
 }()
 log.Println("Connected successfully.")

 log.Println("Starting Ebiten game loop (blocking)...")
 if err := client.Start(); err != nil {
  log.Fatalf("Client exited with error: %v", err)
 }

 log.Println("Client shutdown complete.")
}
```

## Example Callbacks

### Associating/Creating the Connection Entity

Associating an entity to the connection ensures/limits client scope when sending inputs.

```go
func NewConnectionEntityCreate(conn drip.Connection, s drip.Server) (warehouse.Entity, error) {
 serverActiveScenes := s.ActiveScenes()

 if len(serverActiveScenes) == 0 {
  return nil, errors.New("No active scenes to find player in")
 }

 scene := serverActiveScenes[0]
 sto := scene.Storage()

 query := warehouse.Factory.NewQuery().And(components.PlayerSpawnComponent)
 cursor := warehouse.Factory.NewCursor(query, sto)

 var spawn components.PlayerSpawn

 for range cursor.Next() {
  match := components.PlayerSpawnComponent.GetFromCursor(cursor)
  spawn = *match
  break
 }

 return scenes.NewPlayer(spawn.X, spawn.Y, sto)
}
```

### Serialization

`warehouse` provides some helpful tooling to make serialization/deserialization less painful. In this example we
intentionally exclude client meta data components `SpriteBundle` and `SoundBundle`.

```go
func SerializeCallback(scene drip.Scene) ([]byte, error) {
 query := blueprint.Queries.ActionBuffer
 cursor := warehouse.Factory.NewCursor(query, scene.Storage())

 sEntities := []warehouse.SerializedEntity{}

 for range cursor.Next() {

  e, err := cursor.CurrentEntity()
  if err != nil {
   return nil, err
  }

  if !e.Valid() {
   log.Println("skipping invalid", e.Valid())
   continue
  }

  se := e.SerializeExclude(
   client.Components.SpriteBundle,
   client.Components.SoundBundle,
  )

  sEntities = append(sEntities, se)
 }

 serSto := warehouse.SerializedStorage{
  Entities:    sEntities,
  CurrentTick: scene.CurrentTick(),
  Version:     "net",
 }
 stateForJson, err := warehouse.PrepareForJSONMarshal(serSto)
 if err != nil {
  return nil, err
 }
 return json.Marshal(stateForJson)
}
```

### Deserialization

`warehouse` provides some helpful tooling to make serialization/deserialization less painful. In this example we
intentionally add missing client meta data components `SpriteBundle` and `SoundBundle` (since the sever does not send these).
We also use the `ForceSerializedEntityExclude` to avoid touching these client only components when deserializing from
the server.

```go
func Derser(nc coldbrew.NetworkClient, data []byte) error {
 activeScenes := nc.ActiveScenes()
 var scene coldbrew.Scene
 for s := range activeScenes {
  scene = s
  break
 }
 if scene != nil && scene.Ready() {
  storage := scene.Storage()
  if storage != nil {
   var world warehouse.SerializedStorage
   err := json.Unmarshal(data, &world)
   if err != nil {
    log.Printf("NetworkClient Update Error: Failed to unmarshal state (%d bytes): %v", len(data), err)
   } else {

    seen := map[int]struct{}{}

    for _, se := range world.Entities {
     seen[int(se.ID)] = struct{}{}

     en, err := storage.ForceSerializedEntityExclude(
      se, client.Components.SoundBundle,
      client.Components.SpriteBundle,
     )
     if err != nil {
      return err
     }

     err = se.SetValue(en)
     if err != nil {
      return err
     }

     if !en.Table().Contains(client.Components.SpriteBundle) {
      err := en.AddComponentWithValue(client.Components.SpriteBundle, scenes.DEFAULT_PLAYER_SPR_BUNDLE)
      if err != nil {
       return err
      }

      err = en.AddComponentWithValue(client.Components.SoundBundle, scenes.DEFAULT_PLAYER_SND_BUNDLE)
      if err != nil {
       return err
      }
     }

    }

    purge := []warehouse.Entity{}
    query := blueprint.Queries.ActionBuffer
    cursor := scene.NewCursor(query)

    for range cursor.Next() {
     e, _ := cursor.CurrentEntity()
     if _, ok := seen[int(e.ID())]; !ok {
      purge = append(purge, e)
     }
    }

    err := storage.DestroyEntities(purge...)
    if err != nil {
     log.Println(err)
    }

    coldbrew.ForceSetTick(world.CurrentTick)
   }
  } else {
   log.Println("NetworkClient Update Error: Active scene has nil storage.")
  }
 }
 return nil
}
```
