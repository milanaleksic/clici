package server

import (
	"github.com/hashicorp/go-memdb"
	"log"
)

const (
	registrationTable = "registration"
)

type ConnectionID string

func (connectionId *ConnectionID) String() string {
	return string(*connectionId)
}

// Mapping is a wrapper around an in-memory database which will handle current connections
// and metadata related to them
type Mapping struct {
	db *memdb.MemDB
}

// RegisterClient allows registering a certain connection with all requested jobs.
// For a single connection id, this call can be executed multiple times
func (mapping *Mapping) RegisterClient(id ConnectionID, reg Registration) {
	txn := mapping.db.Txn(true)

	if err := txn.Insert(registrationTable, reg); err != nil {
		panic(err)
	}

	txn.Commit()
	log.Println("Connection registered")
}

// UnRegisterClient will remove all mappings from the in-memory DB for a certain connection id
func (mapping *Mapping) UnRegisterClient(id ConnectionID) {
	txn := mapping.db.Txn(true)
	n, err := txn.DeleteAll(registrationTable, "connid", id.String())
	if err != nil {
		log.Printf("Failed when deleting connection records from in-memory DB: %v", err)
	} else {
		log.Printf("Client unregistered, jobs deleted: %d", n)
	}
	txn.Commit()
	log.Println("Connection removed")
}

func (mapping *Mapping) GetAllUniqueJobs() (serverToJobRegistrations map[string][]Registration) {
	txn := mapping.db.Txn(false)
	iterator, err := txn.Get(registrationTable, "jobs")
	if err != nil {
		log.Fatalf("Failed when listing records from in-memory DB: %v", err)
	}
	serverToJobRegistrations = make(map[string][]Registration, 0)
	var iter interface{}
	for {
		iter = iterator.Next()
		if iter == nil {
			break
		}
		reg := iter.(Registration)
		serverToJobRegistrations[reg.ServerLocation] = append(serverToJobRegistrations[reg.ServerLocation], reg)
	}
	return
}

func (mapping *Mapping) FindAllRegisteredConnectionsForServerAndJob(server string, jobName string) (connIds []ConnectionID) {
	txn := mapping.db.Txn(false)
	iterator, err := txn.Get(registrationTable, "jobs", )
	if err != nil {
		log.Fatalf("Failed when listing records from in-memory DB: %v", err)
	}
	connIdsSet := make(map[ConnectionID]bool, 0)
	var iter interface{}
	for {
		iter = iterator.Next()
		if iter == nil {
			break
		}
		reg := iter.(Registration)
		connIdsSet[reg.ConnectionID] = true
	}
	connIds = make([]ConnectionID, 0)
	for connID := range connIdsSet {
		connIds = append(connIds, connID)
	}
	return
}


type Registration struct {
	// ConnectionID is a unique string identifying an active connection from clici client
	ConnectionID   ConnectionID
	// ServerLocation is a location of a Jenkins server some connection is interested in
	ServerLocation string
	// JobName refers to a certain job in the server defined via ServerLocation
	JobName        string
}

// NewMapping creates a single empty Mapping abstraction with ready-for-usage in-memory DB
func NewMapping() *Mapping {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			registrationTable: &memdb.TableSchema{
				Name: registrationTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "ConnectionID"},
								&memdb.StringFieldIndex{Field: "ServerLocation"},
								&memdb.StringFieldIndex{Field: "JobName"},
							},
						},
					},
					"connid": &memdb.IndexSchema{
						Name:   "connid",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "ConnectionID"},
							},
						},
					},
					"jobs": &memdb.IndexSchema{
						Name:   "jobs",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "ServerLocation"},
								&memdb.StringFieldIndex{Field: "JobName"},
							},
						},
					},
				},
			},
		},
	}
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		log.Fatalf("Could not create MemDB, failing entire application: %v", err)
	}
	return &Mapping{
		db: db,
	}
}
