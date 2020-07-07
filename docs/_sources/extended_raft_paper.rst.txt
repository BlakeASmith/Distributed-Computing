Notes from "In Search of an Understandable Consensus Algorithm" (Extended Version)
===================================================================================

- Raft separates Leader Election, Log Replication, and Safety

Features of Raft

* Strong Leader
* Leader Election 
        - leaders are elected based on randomized timers
* Membership changes    
        - *Joint consensus* approach



Raft
-------------------------------

Leader Election
________________________________________

A node can be in one of 3 states (Follower, Candidate, Leader). And 
all nodes start in the **Follower** state. 

A Node becomes a **Candidate** if it does not hear from a Leader. When
a Node becomes a Candidate it will request **votes** from other  nodes. 

Then each other node will reply with it's vote and the **Leader** is 
**elected** if it received votes from the majority of nodes.

Timeout Settings
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

:election timeout: The amount of time a follower waits 
                   before becoming a candidate (randomized)

:Heartbeat timeout: on a regular interval the Leader will send
                    a message to each follower. Upon receiving the 
                    message the follower will reset it's election timeout.

.. note::

        Since a majority of votes is required to be leader it is not 
        possible to have multiple leaders. In the event of a tie, none of the 
        nodes will become Leader. Thus the election timer will again run out 
        and a new election will take place.

Log Replication
_____________________________


Now all changes go through the **Leader** and are noted in it's Log. 

The change remains **uncommited** until the majority of nodes have written 
the change into their logs. (They will send a message to the leader when they 
have done so)

Next, The leader node will notify all follower nodes that the change has been
committed. At which point the cluster has come to a consensus
