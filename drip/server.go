package drip

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/TheBitDrifter/bappa/blueprint"
	"github.com/TheBitDrifter/bappa/blueprint/input"
	"github.com/TheBitDrifter/bappa/table"
	"github.com/TheBitDrifter/bappa/warehouse"
)

type Server interface {
	RegisterScene(name string, width, height int, plan blueprint.Plan, coreSystems []blueprint.CoreSystem) error
	Start() error
	Stop() error
	Broadcast(message []byte) error
	ActiveScenes() []Scene
	GetConnectionEntity(Connection) (warehouse.Entity, bool)
	SetConnectionEntity(Connection, warehouse.Entity) error
	ConsumeAllActions() []bufferedServerActions
	Systems() []ServerSystem
}

type serverImpl struct {
	config       ServerConfig
	scenes       map[string]Scene
	activeScenes []Scene
	listener     net.Listener
	connections  []Connection // active connections
	mutex        sync.RWMutex // Protects scenes, activeScenes, connections, running

	// Network specific fields
	connToEntity map[Connection]warehouse.Entity // Map connection to server entity
	entityMutex  sync.RWMutex                    // Mutex for connToEntity map

	actionQueue actionQueue // Thread-safe queue for incoming actions

	ecsMutex sync.RWMutex // mutex for warehouse.Storage create/destroy operations

	// Control fields
	running bool
	ticker  *time.Ticker
	done    chan bool

	systems []ServerSystem

	deletionChan        chan entityIdentifier
	creationRequestChan chan entityCreationRequest
}

type entityIdentifier struct {
	ID       table.EntryID
	Recycled int
}

type entityCreationRequest struct {
	conn Connection
}

// NewServer creates a new Drip server instance.
func NewServer(config ServerConfig, systems ...ServerSystem) Server {
	deletionChanBufferSize := config.MaxConnections
	if deletionChanBufferSize < 128 {
		deletionChanBufferSize = 128 // Ensure a minimum buffer
	}
	creationChanBufferSize := 32
	return &serverImpl{
		config:       config,
		scenes:       make(map[string]Scene),
		connections:  make([]Connection, 0),
		done:         make(chan bool, 1),
		connToEntity: make(map[Connection]warehouse.Entity),
		actionQueue: actionQueue{
			actions: make([]bufferedServerActions, 0, 128),
		},
		systems:             systems,
		deletionChan:        make(chan entityIdentifier, deletionChanBufferSize),
		creationRequestChan: make(chan entityCreationRequest, creationChanBufferSize),
	}
}

// RegisterScene registers a new scene configuration with the server.
func (s *serverImpl) RegisterScene(
	name string, width, height int, plan blueprint.Plan, coreSystems []blueprint.CoreSystem,
) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.scenes[name]; exists {
		return fmt.Errorf("scene with name '%s' already exists", name)
	}

	schema := table.Factory.NewSchema()
	storage := warehouse.Factory.NewStorage(schema)

	scene := &sceneImpl{
		name:        name,
		width:       width,
		height:      height,
		plan:        plan,
		coreSystems: coreSystems,
		storage:     storage,
		serverMutex: &s.mutex,
		sceneTick:   0,
	}

	if plan != nil {
		if err := plan(width, height, storage); err != nil {
			return fmt.Errorf("failed to execute scene plan: %w", err)
		}
		scene.planExecuted = true
	}

	s.scenes[name] = scene

	// Activate the first registered scene by default
	// For the initial single scene impl we will only have one
	if len(s.activeScenes) == 0 {
		s.activeScenes = append(s.activeScenes, scene)
		log.Printf("Scene '%s' registered and set as active.", name)
	} else {
		log.Printf("Scene '%s' registered.", name)
	}

	return nil
}

// Start begins listening for connections and starts the game loop.
func (s *serverImpl) Start() error {
	s.mutex.Lock()

	if s.running {
		s.mutex.Unlock()
		return fmt.Errorf("server is already running")
	}

	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.mutex.Unlock()
		return fmt.Errorf("failed to start listener on %s: %w", addr, err)
	}

	s.listener = listener
	s.running = true

	// Calculate tick duration and create ticker
	tickDuration := time.Second / time.Duration(s.config.TPS)
	s.ticker = time.NewTicker(tickDuration)

	// Unlock before starting goroutines
	s.mutex.Unlock()

	go s.acceptConnections()
	go s.gameLoop()

	log.Printf("Server started on port %d with TPS: %d", s.config.Port, s.config.TPS)
	return nil
}

// Stop gracefully shuts down the server.
func (s *serverImpl) Stop() error {
	s.mutex.Lock()
	close(s.deletionChan)
	close(s.creationRequestChan)

	if !s.running {
		s.mutex.Unlock()
		return fmt.Errorf("server is not running")
	}

	log.Println("Stop: Initiating server shutdown...")

	s.running = false

	//  Signal game loop to stop
	select {
	case s.done <- true:
		log.Println("Stop: Sent stop signal to game loop.")
	default:
		log.Println("Stop: Stop signal already sent or channel full.")
	}

	if s.ticker != nil {
		s.ticker.Stop()
		log.Println("Stop: Ticker stopped.")
	}

	// Close the listener to prevent new connections and unblock Accept()
	var listenerErr error
	if s.listener != nil {
		log.Println("Stop: Closing listener...")
		listenerErr = s.listener.Close()
		if listenerErr != nil {
			log.Printf("Stop: Error closing listener: %v", listenerErr)
		} else {
			log.Println("Stop: Listener closed.")
		}
	}

	connsToClose := make([]Connection, len(s.connections))
	copy(connsToClose, s.connections)
	log.Printf("Stop: Copied %d connections to close.", len(connsToClose))

	s.connections = make([]Connection, 0)

	//  Release the mutex *before* closing connections
	s.mutex.Unlock()

	log.Println("Stop: Closing connections...")
	var wg sync.WaitGroup
	for _, conn := range connsToClose {
		wg.Add(1)
		go func(c Connection) {
			defer wg.Done()
			addr := c.Address()
			log.Printf("Stop: Closing connection %s", addr)
			if err := c.Close(); err != nil {
				log.Printf("Stop: Error closing connection %s: %v", addr, err)
			}
		}(conn)
	}
	wg.Wait()

	log.Println("Stop: All connections closed. Server shutdown complete.")
	return listenerErr
}

// Broadcast sends a message to all currently connected clients.
func (s *serverImpl) Broadcast(message []byte) error {
	s.mutex.RLock()

	connsToSend := make([]Connection, len(s.connections))
	copy(connsToSend, s.connections)

	s.mutex.RUnlock()

	if len(connsToSend) == 0 {
		return nil // No clients connected.
	}

	for _, conn := range connsToSend {
		go func(c Connection, msg []byte) {
			if err := c.Send(msg); err != nil {
				log.Println(err)
			}
		}(conn, message)
	}
	// Return immediately - does not wait for sends to complete, currently.
	// Eventually may launch a routine and check for err via channel, but really
	// don't want to block whilst waiting on unresponsive clients
	return nil
}

func (s *serverImpl) ActiveScenes() []Scene {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	scenes := make([]Scene, len(s.activeScenes))
	copy(scenes, s.activeScenes)
	return scenes
}

func (s *serverImpl) GetConnectionEntity(c Connection) (warehouse.Entity, bool) {
	s.entityMutex.RLock()
	defer s.entityMutex.RUnlock()
	en, ok := s.connToEntity[c]
	return en, ok
}

func (s *serverImpl) SetConnectionEntity(c Connection, e warehouse.Entity) error {
	s.entityMutex.Lock()
	defer s.entityMutex.Unlock()
	s.connToEntity[c] = e

	return nil
}

func (s *serverImpl) ConsumeAllActions() []bufferedServerActions {
	return s.actionQueue.ConsumeAll()
}

func (s *serverImpl) Systems() []ServerSystem {
	return s.systems
}

func (s *serverImpl) acceptConnections() {
	log.Println("Acceptor: Goroutine started.")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Acceptor: Listener closed, exiting goroutine.")
				return
			}
			s.mutex.RLock()
			running := s.running
			s.mutex.RUnlock()
			if !running {
				log.Println("Acceptor: Server not running and Accept error occurred, exiting goroutine.")
				return
			}
			log.Printf("Acceptor: Error accepting connection: %v. Continuing...", err)
			continue
		}

		s.mutex.Lock()
		if len(s.connections) >= s.config.MaxConnections {
			log.Printf("Acceptor: Max connections (%d) reached. Rejecting %s", s.config.MaxConnections, conn.RemoteAddr())
			conn.Close()
			s.mutex.Unlock()
			continue
		}

		clientConn := NewConnection(conn)
		s.connections = append(s.connections, clientConn)
		log.Printf("Acceptor: Connection %s accepted. Total: %d", conn.RemoteAddr(), len(s.connections))
		s.mutex.Unlock()

		go s.handleConnection(clientConn)
	}
}

func (s *serverImpl) gameLoop() {
	log.Println("GameLoop: Goroutine started.")
	defer log.Println("GameLoop: Goroutine finished.")

	for {
		select {
		case <-s.done:
			log.Println("GameLoop: Stop signal received. Exiting.")
			return
		case <-s.ticker.C:
			if err := s.update(); err != nil {
				log.Fatalf("GameLoop: Error during update: %v", err)
			}
		}
	}
}

// isUseOfClosedConnErr checks for common network connection closed errors.
func isUseOfClosedConnErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	// Fallback check for older Go versions or specific OS errors
	return err.Error() == "use of closed network connection"
}

// update executes a single server tick. It processes entity creation/deletion requests,
// runs server/core systems, serializes state,
// and broadcasts the state to active clients.
func (s *serverImpl) update() error {
DrainCreationChannel:
	for {
		select {
		case req := <-s.creationRequestChan:
			log.Printf("[Update] Processing creation request for %s", req.conn.Address())

			var createdEntity warehouse.Entity
			var creationErr error

			// Call the game-specific creation callback, protected by the ECS mutex.
			s.ecsMutex.Lock()
			createdEntity, creationErr = Callbacks.NewConnectionCreateEntity(req.conn, s)
			s.ecsMutex.Unlock()

			if creationErr != nil {
				log.Printf("[Update] Callback NewConnectionCreateEntity failed for %s: %v", req.conn.Address(), creationErr)
				return creationErr
			}

			s.mutex.Lock()
			s.SetConnectionEntity(req.conn, createdEntity)
			log.Printf("Associated connection %s with player entity ID %d", req.conn.Address(), createdEntity.ID())

			assignMsg := AssignEntityIDMessage{
				Type:     AssignEntityIDMessageType,
				EntityID: createdEntity.ID(),
			}
			jsonData, err := json.Marshal(assignMsg)
			if err != nil {
				log.Printf("Error marshalling AssignEntityIDMessage for %s: %v", req.conn.Address(), err)
				return nil
			}

			log.Printf("Sending EntityID %d to %s", createdEntity.ID(), req.conn.Address())
			err = req.conn.Send(jsonData)
			if err != nil {
				log.Printf("Error sending AssignEntityIDMessage to %s: %v", req.conn.Address(), err)
			}
			s.mutex.Unlock()
			log.Printf("[Update] Callback NewConnectionCreateEntity succeeded for %s, created entity %d", req.conn.Address(), createdEntity.ID())
		default:
			break DrainCreationChannel
		}
	}

DrainDeletionChannel:
	for {
		select {
		case req := <-s.deletionChan:
			log.Printf("[Update] Processing deletionChan request for %d", req.ID)

			activeScene := s.ActiveScenes()[0]

			// Call the game-specific creation callback, protected by the ECS mutex.
			s.ecsMutex.Lock()
			sto := activeScene.Storage()
			en, err := sto.Entity(int(req.ID))
			if err != nil {
				s.ecsMutex.Unlock()
				return err
			}
			delError := sto.DestroyEntities(en)
			s.ecsMutex.Unlock()

			if delError != nil {
				s.ecsMutex.Unlock()
				return delError
			}

		default:
			break DrainDeletionChannel
		}
	}

	// Systems
	s.ecsMutex.Lock()
	for _, serverSys := range s.Systems() {
		if err := serverSys.Run(s); err != nil {
			log.Printf("[Update] Error running server system %T: %v", serverSys, err)
			s.ecsMutex.Unlock()
			return err
		}
	}

	tps := s.config.TPS
	deltaTime := 1.0 / float64(tps)
	activeScenes := s.ActiveScenes()

	for _, scene := range activeScenes {
		// Core systems iterate and modify ECS state. Protected by the outer ecsMutex.
		for _, system := range scene.CoreSystems() {
			if err := system.Run(scene, deltaTime); err != nil {
				log.Printf("[Update] Error running core system %T on scene '%s': %v", system, scene.Name(), err)
				s.ecsMutex.Unlock()
				return err
			}
		}
		scene.IncrementTick()
	}
	s.ecsMutex.Unlock()

	// Serialize State
	var state []byte
	var stateErr error
	if len(activeScenes) > 0 {
		s.ecsMutex.RLock()
		state, stateErr = Callbacks.Serialize(activeScenes[0])
		s.ecsMutex.RUnlock()
	}

	if stateErr != nil {
		log.Printf("[Update] Error serializing state: %v", stateErr)
		return stateErr
	} else if state != nil {
		// Broadcast uses the current s.connections list, which is modified
		// concurrently by handleConnection. Broadcast needs to handle this safely.
		if err := s.Broadcast(state); err != nil {
			log.Printf("[Update] Error occurred during broadcast phase: %v", err)
		}
	}

	return nil
}

func (s *serverImpl) handleConnection(conn Connection) {
	connAddr := conn.Address()
	log.Printf("Handler [%s]: Goroutine started.", connAddr)

	var associatedEntity warehouse.Entity = nil

	defer func() {
		s.mutex.RLock()
		associatedEntity = s.connToEntity[conn]
		s.mutex.RUnlock()

		log.Printf("Handler [%s]: Initiating cleanup.", connAddr)

		if err := conn.Close(); err != nil {
			if !errors.Is(err, net.ErrClosed) && !isUseOfClosedConnErr(err) {
				log.Printf("Handler [%s]: Error during conn.Close(): %v", connAddr, err)
			}
		} else {
			log.Printf("Handler [%s]: Network connection closed.", connAddr)
		}
		if associatedEntity != nil {
			s.entityMutex.Lock()
			currentEntity, exists := s.connToEntity[conn]
			if exists && currentEntity == associatedEntity {
				delete(s.connToEntity, conn)
				log.Printf("Handler [%s]: Removed connection mapping for entity %d.", connAddr, associatedEntity.ID())
			} else if exists {
				log.Printf("Handler [%s]: Warning! Stale/mismatched entity mapping found during cleanup (found %v, expected %v).", connAddr, currentEntity, associatedEntity)
			}
			s.entityMutex.Unlock()

			entityID := associatedEntity.ID()
			recycledCount := associatedEntity.Recycled()
			select {
			case s.deletionChan <- entityIdentifier{ID: entityID, Recycled: recycledCount}:
				log.Printf("Handler [%s]: Queued entity ID %d (Recycled: %d) for deletion.", connAddr, entityID, recycledCount)
			default:
				log.Printf("Handler [%s]: Warning! Deletion channel full. Failed to queue entity ID %d.", connAddr, entityID)
			}
		} else {
			log.Printf("Handler [%s]: No entity fully associated, skipping deletion queue.", connAddr)
		}

		removed := false
		removeTimer := time.NewTimer(1 * time.Second)
		select {
		case <-func() chan struct{} { // Anonymous func closure to use select for lock attempt
			done := make(chan struct{})
			go func() {
				s.mutex.Lock()
				// Use swap-and-pop for efficient removal from slice
				for i, c := range s.connections {
					if c == conn {
						lastIdx := len(s.connections) - 1
						s.connections[i] = s.connections[lastIdx]
						s.connections[lastIdx] = nil
						s.connections = s.connections[:lastIdx]
						removed = true
						break
					}
				}
				s.mutex.Unlock()
				close(done)
			}()
			return done
		}():

			// Lock acquired and removal attempt completed.
			removeTimer.Stop()
			if removed {
				log.Printf("Handler [%s]: Removed connection from active list.", connAddr)
			} else {
				log.Printf("Handler [%s]: Connection %s not found in active list during cleanup.", connAddr, connAddr)
			}
		case <-removeTimer.C:
			log.Printf("Handler [%s]: Warning! Timeout acquiring lock to remove connection %s from active list.", connAddr, connAddr)
			break
		}
		log.Printf("Handler [%s]: Cleanup finished. Goroutine exiting.", connAddr)
	}()

	request := entityCreationRequest{
		conn: conn,
	}

	creationRequestTimeout := 2 * time.Second

	select {
	case s.creationRequestChan <- request:
		log.Printf("Handler [%s]: Entity creation request sent.", connAddr)
	case <-time.After(creationRequestTimeout):
		log.Printf("Handler [%s]: Timeout sending entity creation request after %v. Closing connection.", connAddr, creationRequestTimeout)
		return // Exit, triggers cleanup
	}

	log.Printf("Handler [%s]: Starting receive loop.", connAddr)
	readTimeout := 10 * time.Second
	msgProcessingTimeout := 2 * time.Second
	for {
		s.mutex.RLock()
		associatedEntity = s.connToEntity[conn]
		s.mutex.RUnlock()

		// Receive message
		messageBytes, err := conn.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || isUseOfClosedConnErr(err) {
				log.Printf("Handler [%s]: Connection closed (EOF or closed error).", connAddr)
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				log.Printf("Handler [%s]: Read timeout after %v of inactivity.", connAddr, readTimeout)
			} else {
				log.Printf("Handler [%s]: Unhandled error receiving message: %v.", connAddr, err)
			}
			return
		}

		msgProcessed := make(chan bool, 1)

		go func(msg []byte) {
			s.mutex.RLock()
			associatedEntity = s.connToEntity[conn]
			s.mutex.RUnlock()

			if associatedEntity == nil {
				log.Printf("Handler [%s]: Internal error! associatedEntity is nil in receive loop. Discarding.", connAddr)
				msgProcessed <- false
				return
			}
			s.ecsMutex.RLock()
			entityID := associatedEntity.ID()
			recycledCount := associatedEntity.Recycled()
			s.ecsMutex.RUnlock()

			var actionMsg input.ClientActionMessage
			if err := json.Unmarshal(msg, &actionMsg); err != nil {
				log.Printf("Handler [%s]: Error unmarshalling input message: %v. Discarding.", connAddr, err)
				msgProcessed <- false
				return
			}

			queuedActions := bufferedServerActions{
				TargetEntityID: entityID, Recycled: recycledCount,
				ReceiverIndex: actionMsg.ReceiverIndex, Actions: actionMsg.Actions,
			}
			s.actionQueue.mu.Lock()
			s.actionQueue.actions = append(s.actionQueue.actions, queuedActions)
			s.actionQueue.mu.Unlock()
			msgProcessed <- true
		}(messageBytes)

		// Wait for processing
		select {
		case success := <-msgProcessed:
			if !success {
				log.Printf("Handler [%s]: Message processing failed.", connAddr)
			}
		case <-time.After(msgProcessingTimeout):
			log.Printf("Handler [%s]: Message processing timed out after %v.", connAddr, msgProcessingTimeout)
		}
	}
}
