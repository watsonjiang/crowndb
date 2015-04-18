#CrownDB
-----------
CrownDB is a memory based storage system. 


##Design
There're elementry concepts in CrownDB.

keyspace - a namespace for data. the data in different keyspace is isolated both in logically and pysically.

bucket - a logical data holder. generally, keyspace splits into several buckets.

tablet - a pysical data holder. a bucket need to be hold at least one tablet, a tablet can hold more then one buckets

server - pysical machines.

##Folder structure
ha  - the hash function
log - the crowndb logging interface.(for bridging to different logging imp)
mem - a module implements a mem pool, used by se.
re - replication engine, the module replicate data to other server.
se - storage engine, the module hold the data.
cfs - module which maintains the config data
pr - the protocol

