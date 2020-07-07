Summary of Distributed Systems for Fun and Profit Chapter 4
=============================================================
See the full chapter here at http://book.mixu.net/distsys/single-page.html

**Chapter 4** discusses the **replication problem** and various strategies for 
dealing with it. These strategies are grouped into two major categories:

:Single Copy Systems: prevent divergence by ensuring only a single copy of the 
                      data is active at any given time. Ensure replicas are in agreement
                      such that when one fails, another can replace it without issue.

:Multi-Master Systems: risk divergence by maintaining multiple active copies of the data 
                       at the same time. Availability will be improved, but at the cost of 
                       preventing divergence


The author goes on to discuss **Primary/Backup replication**, **Two Phase Commit**, and **Partition Tolerant
Consensus Algorithms** such as **Raft**.

**Primary/Backup replication** is the simplest of the bunch and comes in *syncronous* and *asyncronous* flavors. 
In this replication scheme, all updates are perfomed at the primary (leader) data store. The updates are then
pushed to the backup (follower) replicas.

The following sequence diagrams show the Asyncronous and Syncronous versions of this scheme.

.. raw:: html

        <pre> 
        Sync Primary/Backup: 1 message per replica
                                                                       
        ┌─┐             ,.-^^-._           ,.-^^-._           ,.-^^-._ 
        ║"│            |-.____.-|         |-.____.-|         |-.____.-|
        └┬┘            |        |         |        |         |        |
        ┌┼┐            |        |         |        |         |        |
         │             |        |         |        |         |        |
        ┌┴┐            '-.____.-'         '-.____.-'         '-.____.-'
      Client            Primary            Backup1            Backup2  
        │   write(DATA)    │                  │                  │     
        │ ────────────────>│                  │                  │     
        │                  │                  │                  │     
        │           ╔══════╧══════╗           │                  │     
        │           ║writes data ░║           │                  │     
        │           ╚══════╤══════╝           │                  │     
        │      return      │                  │                  │     
        │ <────────────────│                  │                  │     
        │                  │                  │                  │     
        │                  │   wrtie(DATA)    │                  │     
        │                  │─────────────────>│                  │     
        │                  │                  │                  │     
        │                  │            write(DATA)              │     
        │                  │────────────────────────────────────>│     
        │                  │                  │                  │     
        │                  │             ╔════╧══════════════════╧════╗
        │                  │             ║writes data                ░║
      Client            Primary          ╚════════════════════════════╝
        ┌─┐             ,.-^^-._           ,.-^^-._           ,.-^^-._ 
        ║"│            |-.____.-|         |-.____.-|         |-.____.-|
        └┬┘            |        |         |        |         |        |
        ┌┼┐            |        |         |        |         |        |
         │             |        |         |        |         |        |
        ┌┴┐            '-.____.-'         '-.____.-'         '-.____.-'





        Sync Primary/Backup: 2 messages per replica
                                                                       
        ┌─┐             ,.-^^-._           ,.-^^-._           ,.-^^-._ 
        ║"│            |-.____.-|         |-.____.-|         |-.____.-|
        └┬┘            |        |         |        |         |        |
        ┌┼┐            |        |         |        |         |        |
         │             |        |         |        |         |        |
        ┌┴┐            '-.____.-'         '-.____.-'         '-.____.-'
      Client            Primary            Backup1            Backup2  
        │   write(DATA)    │                  │                  │     
        │ ────────────────>│                  │                  │     
        │                  │                  │                  │     
        │           ╔══════╧══════╗           │                  │     
        │           ║writes data ░║           │                  │     
        │           ╚══════╤══════╝           │                  │     
        │                  │   wrtie(DATA)    │                  │     
        │                  │─────────────────>│                  │     
        │                  │                  │                  │     
        │                  │            write(DATA)              │     
        │                  │────────────────────────────────────>│     
        │                  │                  │                  │     
        │                  │             ╔════╧══════════════════╧════╗
        │                  │             ║writes data                ░║
        │                  │             ╚════╤══════════════════╤════╝
        │                  │       ACK        │                  │     
        │                  │<─────────────────│                  │     
        │                  │                  │                  │     
        │                  │                ACK                  │     
        │                  │<────────────────────────────────────│     
        │                  │                  │                  │     
        │      return      │                  │                  │     
        │ <────────────────│                  │                  │     
      Client            Primary            Backup1            Backup2  
        ┌─┐             ,.-^^-._           ,.-^^-._           ,.-^^-._ 
        ║"│            |-.____.-|         |-.____.-|         |-.____.-|
        └┬┘            |        |         |        |         |        |
        ┌┼┐            |        |         |        |         |        |
         │             |        |         |        |         |        |
        ┌┴┐            '-.____.-'         '-.____.-'         '-.____.-' 


        </pre>

The next step up is **Two Phase Commit**. It works alot like *syncronous Primary/Backup* in the beginning.
The difference is that the replicas will not write to their "real" logs immediatley. Instead they will write to 
a temporary (*write-ahead*) log. Each backup will then vote on whether to commit the entry or abort it. Once the votes 
are received at the Primary, a decition will be made and the backups will be updated. Updating the diagram from before we get:

.. raw:: html

        <pre>

                         2PC:4 messages per replica                        
                                                                           
        ┌─┐             ,.-^^-._           ,.-^^-._           ,.-^^-._     
        ║"│            |-.____.-|         |-.____.-|         |-.____.-|    
        └┬┘            |        |         |        |         |        |    
        ┌┼┐            |        |         |        |         |        |    
         │             |       |         |        |         |        |    
        ┌┴┐            '-.____.-'         '-.____.-'         '-.____.-'    
      Client            Primary            Backup1            Backup2      
        │   write(DATA)    │                  │                  │         
        │ ────────────────>│                  │                  │         
        │                  │                  │                  │         
        │     ╔════════════╧════════════╗     │                  │         
        │     ║writes data to temp log ░║     │                  │         
        │     ╚════════════╤════════════╝     │                  │         
        │                  │   wrtie(DATA)    │                  │         
        │                  │─────────────────>│                  │         
        │                  │                  │                  │         
        │                  │            write(DATA)              │         
        │                  │────────────────────────────────────>│         
        │                  │                  │                  │         
        │                  │             ╔════╧══════════════════╧════╗    
        │                  │             ║writes data to temp log    ░║    
        │                  │             ╚════╤══════════════════╤════╝    
        │                  │vote COMMIT/ABORT │                  │         
        │                  │<─────────────────│                  │         
        │                  │                  │                  │         
        │                  │         vote COMMIT/ABORT           │         
        │                  │<────────────────────────────────────│         
        │                  │                  │                  │         
        │        ╔═════════╧══════════╗       │                  │         
        │        ║makes a decition   ░║       │                  │         
        │        ║based on the votes  ║       │                  │         
        │        ╚═════════╤══════════╝       │                  │         
        │                  │  COMMIT/ABORT    │                  │         
        │                  │─────────────────>│                  │         
        │                  │                  │                  │         
        │                  │            COMMIT/ABORT             │         
        │                  │────────────────────────────────────>│         
        │                  │                  │                  │         
        │                  │         ╔════════╧══════════════════╧════════╗
        │                  │         ║writes to the "real" log or aborts ░║
        │                  │         ╚════════╤══════════════════╤════════╝
        │                  │       ACK        │                  │         
        │                  │<─────────────────│                  │         
        │                  │                  │                  │         
        │                  │                ACK                  │         
        │                  │<────────────────────────────────────│         
        │                  │                  │                  │         
        │      ╔═══════════╧═══════════╗      │                  │         
        │      ║writes to the "real"  ░║      │                  │         
        │      ║log or aborts          ║      │                  │         
        │      ╚═══════════╤═══════════╝      │                  │         
        │      return      │                  │                  │         
        │ <────────────────│                  │                  │         
      Client            Primary            Backup1            Backup2      
        ┌─┐             ,.-^^-._           ,.-^^-._           ,.-^^-._     
        ║"│            |-.____.-|         |-.____.-|         |-.____.-|    
        └┬┘            |        |         |        |         |        |    
        ┌┼┐            |        |         |        |         |        |    
         │             |        |         |        |         |        |    
        ┌┴┐            '-.____.-'         '-.____.-'         '-.____.-'     


        </pre>

Finally, the autor discusses partion tolerant algorithms like *Raft* and *Paxos*. These algorthms 
add partion tolerance to *2PC* by requiring agreement among only a **majority** of nodes (as apposed
to total consensus). By doing so, they ensure that only one (leader) replica is active during a partion
(as the smaller partition would not be able to reach a majority vote). Disagreements will may 
prevent the algorithm from making progress, but single-copy consistency will not be violated.

Questions
____________________

After reading through this chapter the following questions come to mind:


* When is it sufficient to use 2PC, or even Primary/Backup, instead of an algroithm such as Raft or Paxos?
* What are examples of *Multi-Master* style replication strategies? 
        * The author mentioned Mulit-Master systems, but did not go into any details or give expamples

Research
_____________________

Since it wasn't really addressed in the artice, I did some reading on Multi-Master repliaction schemes from the following 
sources:

1. https://en.wikipedia.org/wiki/Multi-master_replication
2. https://dzone.com/articles/pros-and-cons-of-mysql-replication-types
3. https://dev.mysql.com/doc/refman/5.7/en/mysql-cluster-replication-multi-master.html
4. https://blogs.oracle.com/jsmyth/circular-replication-in-mysql


Multi-Master systems themselves can come in a few configurations;

you can have **Circular Replication**, **Master-Master Replication with N masters**, and  **Master Group Replication (MGR)**

:Circular Replication:
        circular replication is exactly what it sounds like. Every node is a master, and every master replicates to exactly one 
        other master. So if there were 4 nodes in my cluster labeled A, B, C, D the replication would flow as 


        A -> B -> C -> D -> A


        Then, a client could perform a write at any node, and it would be replicated to all others since they are all 
        connected in this circle. 

        
        There's an obvious problem with this... If any one node goes down the circle breaks. This means that failure at one node 
        means failure of the entire system, making this less reliable than just having a single server. The benifit is **write scalability**, since 
        writes are able to be performed at any node.


:Master-Master Replication with N Masters:
        In this scheme we habe N masters and M followers. We get better write-scalability as we have more master nodes we can write too. Additionally the 
        Master nodes can be spaced out geographically to provide lower latencies in more places. The Masters can all replicate to eachother and the followers.
        Writes are accepted at any Master node and replicated to all other nodes.

:Master Group Replication:
        In Master Group Replication, the set of all nodes is split into a number of groups. Then each group will have it's own master and 
        leader election process within that group. Writes are accepted at any of the Master nodes, but the replication works a little 
        differently. Instead of the single Master (the one that received the write request) replicating the data across all nodes, It makes a 
        write request to all of the other Master nodes. Then each Master handles replicating the data to it's own followers.
        













