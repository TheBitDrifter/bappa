package coldbrew

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"sync"

	"github.com/TheBitDrifter/bappa/drip"
	"github.com/hajimehoshi/ebiten/v2"
)

var errNotAssignEntityIDMessage = errors.New("message is not AssignEntityIDMessage")

// NetworkClient extends the base Client interface with networking capabilities.
type NetworkClient interface {
	Client

	// Connect establishes a connection to a drip server.
	Connect(address string) error

	// Disconnect closes the connection to the server.
	Disconnect() error

	// Send transmits data to the server.
	Send(data []byte) error

	// IsConnected returns true if the client is currently connected.
	IsConnected() bool

	SetDeserCallback(func(NetworkClient, []byte) error)

	AssociatedEntityID() (int, bool)
}

// networkClientImpl is the concrete implementation of NetworkClient.
// It embeds the standard client implementation and adds networking logic with drip
type networkClientImpl struct {
	*clientImpl             // Embed the standard client implementation
	dripClient  drip.Client // The underlying drip network client

	// TODO: Consider atomic bool here
	isConnected bool
	stateMutex  sync.RWMutex
	// end of todo -----
	//
	deserCallback func(nc NetworkClient, data []byte) error

	associatedEntityID    int // Server-assigned entity ID.
	hasAssociatedEntityID bool
	assocEntityIDMutex    sync.RWMutex

	hasReceivedStateOnce bool
}

// NewNetworkClient creates a new network-enabled client.
func NewNetworkClient(baseResX, baseResY, maxSpritesCached, maxSoundsCached, maxScenesCached int, embeddedFS fs.FS) NetworkClient {
	// Create the base client implementation
	baseClient := newClientImplBase(baseResX, baseResY, maxSpritesCached, maxSoundsCached, maxScenesCached, embeddedFS)

	// Create the network client wrapper
	nc := &networkClientImpl{
		clientImpl:            baseClient,
		dripClient:            drip.NewClient(10),
		isConnected:           false,
		associatedEntityID:    0,
		hasAssociatedEntityID: false,
	}

	return nc
}

func (cli *networkClientImpl) Start() error {
	if len(cli.loadingScenes) == 0 {
		cli.loadingScenes = append(cli.loadingScenes, defaultLoadingScene)
	}

	err := ebiten.RunGame(cli)
	if err != nil {
		return err
	}
	return nil
}

// Connect establishes the connection using the embedded drip client.
func (nc *networkClientImpl) Connect(address string) error {
	nc.stateMutex.Lock()
	if nc.isConnected {
		nc.stateMutex.Unlock()
		log.Printf("NetworkClient: Already connected to %s", address)
		return nil
	}

	// Unlock before potentially blocking network call
	nc.stateMutex.Unlock()

	log.Printf("NetworkClient: Connecting to %s...", address)
	err := nc.dripClient.Connect(address)
	if err != nil {
		log.Printf("NetworkClient: Connection failed: %v", err)
		return err
	}

	nc.stateMutex.Lock()
	nc.isConnected = true
	nc.stateMutex.Unlock()

	log.Printf("NetworkClient: Successfully connected to %s", address)

	return nil
}

// Disconnect closes the connection.
func (nc *networkClientImpl) Disconnect() error {
	nc.stateMutex.Lock()
	if !nc.isConnected {
		nc.stateMutex.Unlock()
		log.Println("NetworkClient: Not connected.")
		return nil
	}
	nc.isConnected = false
	nc.stateMutex.Unlock() // Unlock before closing

	log.Println("NetworkClient: Disconnecting...")
	err := nc.dripClient.Disconnect()
	if err != nil {
		log.Printf("NetworkClient: Error during disconnect: %v", err)
	} else {
		log.Println("NetworkClient: Disconnected.")
	}
	return err
}

// Send transmits data via the drip client.
func (nc *networkClientImpl) Send(data []byte) error {
	nc.stateMutex.RLock()
	connected := nc.isConnected
	nc.stateMutex.RUnlock()

	if !connected {
		return errors.New("NetworkClient: not connected")
	}
	return nc.dripClient.Send(data)
}

// IsConnected returns the current connection status.
func (nc *networkClientImpl) IsConnected() bool {
	nc.stateMutex.RLock()
	defer nc.stateMutex.RUnlock()
	return nc.isConnected
}

// Update processes network messages and runs client logic.
func (nc *networkClientImpl) Update() error {
	if nc.IsConnected() {
		var latestStateData []byte = nil
		messageBuffer := nc.dripClient.Buffer()

		// Process all buffered messages.
		for len(messageBuffer) > 0 {
			msgData := <-messageBuffer

			// Attempt to process as the special ID message first.
			err := nc.tryProcessAssignEntityID(msgData)
			if err == nil {
				continue // ID message processed, skip to next buffer item.
			} else if errors.Is(err, errNotAssignEntityIDMessage) {
				latestStateData = msgData // Not the ID message, keep as potential state.
			} else {
				// Error during check itself.
				log.Printf("NetworkClient Update: Error checking for AssignEntityID type: %v", err)
				latestStateData = msgData
			}
		}

		// Process the last non-ID message using the general callback.
		if latestStateData != nil {
			if nc.deserCallback != nil {
				if deserErr := nc.deserCallback(nc, latestStateData); deserErr != nil {
					log.Printf("NetworkClient Update: Deserialization callback error: %v", deserErr)
				}
			} else {
				log.Println("NetworkClient Update: Received unhandled non-AssignEntityID message; no deserCallback set.")
			}
			nc.hasReceivedStateOnce = true
		}
	}
	// Run standard client update logic.
	if err := sharedClientUpdate(nc); err != nil {
		log.Printf("Error during shared client update: %v", err)
		return err
	}
	return nil
}

func (cli *networkClientImpl) run() error {
	if !cli.hasReceivedStateOnce {
		return nil
	}
	for _, globalClientSystem := range cli.globalClientSystems {
		err := globalClientSystem.Run(cli)
		if err != nil {
			return err
		}
	}
	loadingScenes := cli.loadingScenes
	for activeScene := range cli.ActiveScenes() {
		cameraReady := true
		cameras := cli.ActiveCamerasFor(activeScene)
		for _, cam := range cameras {
			if !cam.Ready(cli) {
				cameraReady = false
			}
		}
		if !cameraReady || !activeScene.Ready() {
			if len(loadingScenes) > 0 {
				loadingScene := loadingScenes[0]
				for _, coreSys := range loadingScene.CoreSystems() {
					err := coreSys.Run(loadingScene, 1.0/float64(ClientConfig.tps))
					if err != nil {
						return err
					}
				}
				for _, clientSys := range loadingScene.ClientSystems() {
					err := clientSys.Run(cli, loadingScene)
					if err != nil {
						return err
					}
				}
			}
		}
		if activeScene.Ready() {
			for _, coreSys := range activeScene.CoreSystems() {
				err := coreSys.Run(activeScene, 1.0/float64(ClientConfig.tps))
				if err != nil {
					return err
				}
			}
			for _, clientSys := range activeScene.ClientSystems() {
				err := clientSys.Run(cli, activeScene)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Set deserialization function
func (nc *networkClientImpl) SetDeserCallback(cb func(NetworkClient, []byte) error) {
	nc.deserCallback = cb
}

func (nc *networkClientImpl) AssociatedEntityID() (int, bool) {
	nc.assocEntityIDMutex.RLock()
	defer nc.assocEntityIDMutex.RUnlock()
	return int(nc.associatedEntityID), nc.hasAssociatedEntityID
}

// tryProcessAssignEntityID attempts to decode data as an AssignEntityIDMessage.
// If successful, it stores the ID and returns nil. Otherwise, returns an error.
func (nc *networkClientImpl) tryProcessAssignEntityID(data []byte) error {
	var msg drip.AssignEntityIDMessage
	// Attempt to unmarshal using the specific message structure.
	if err := json.Unmarshal(data, &msg); err != nil {
		return errNotAssignEntityIDMessage // Likely wrong structure.
	}
	if msg.Type != drip.AssignEntityIDMessageType {
		return errNotAssignEntityIDMessage // Correct structure, wrong type.
	}

	nc.assocEntityIDMutex.Lock()
	nc.associatedEntityID = int(msg.EntityID)
	nc.hasAssociatedEntityID = true
	nc.assocEntityIDMutex.Unlock()
	log.Printf("NetworkClient: Stored associated Entity ID: %d", msg.EntityID)
	return nil
}

func (cli *networkClientImpl) Draw(image *ebiten.Image) {
	sharedDraw(cli, image)
}
