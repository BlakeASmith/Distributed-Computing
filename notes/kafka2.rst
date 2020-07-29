
Kafka basics
----------------

- System for logging events
- NOT A QUEUE, it's a log
        - it stores things

- Topics
        - events, key-value pairs
        - live in topic until some specified expiry time
        - can be the system of record, but is usually temporary
        - log partitioned over multiple machines (brokers)
                - consistent hashing

- Ordering
        - ordering is by key
                - no global ordering based on time
                - usually doesnt matter though, ordering by key is usually good enough

- Replication
        - usually 3
        - elects a leader 
                - producers talk to leader
                - comsumers also talk to the leader
        - can have multiple consumers on a topic

- everything is a producer or a consumer, that's all kafka knows about.
  things like kafka streams are build on top

- events are immutable
- partitions are not immutable
        - replicating mutable objects is painfull

- Conroller

        - a special broker that is elected as the controller
                - monitors other brokers
                - when a broker fails the controller medates the elections per parition
                        - each broker is a folower for some partitions and a leader of others
                        - controller will update other brokers about new leaders per topic and partitions

        - followers ask the leader for new writes
                - writes are not pushed to followers
                - what determines a successful write is determined by the application
                - can only consume data that has been fully replicated

In the keynote the speaker brought voulanteers to act out the operations of a kafka cluster

        - leader can't expose items until they are replicated
        - dumb brokers are good

- consumer groups

        - one consumer is the leader of the group
        - one broker is in charge of that consumer group
        - first member of group is the keader
        - leader sends heartbeats to broker
        - first consumer gets all the partitions
        - leader rebalances when new members join the group
        - the other members will get partition assignments on the next heartbeat
