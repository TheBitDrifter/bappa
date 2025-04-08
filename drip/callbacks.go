package drip

import (
	"errors"
	"log"

	"github.com/TheBitDrifter/bappa/warehouse"
)

var Callbacks = callbacks{
	NewConnectionCreateEntity: DefaultNewConnectionCreateEntity,
	Serialize:                 DefaultSerializeCallback,
}

type callbacks struct {
	NewConnectionCreateEntity func(conn Connection, s Server) (warehouse.Entity, error)
	Serialize                 func(scene Scene) ([]byte, error)
}

func DefaultNewConnectionCreateEntity(conn Connection, s Server) (warehouse.Entity, error) {
	log.Println("implement your own entity assoc callback for new conns please")
	return nil, nil
}

func DefaultSerializeCallback(scene Scene) ([]byte, error) {
	return nil, errors.New("please implement serializer callback")
}
