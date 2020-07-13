Summary of Chapter 5
==============================

See the full book here: http://book.mixu.net/distsys/single-page.html#eventual


Chapter 4 of this book covered replication techniques where **Single-copy consistency** (replicas must not
diverge) is required. 
It's possible to do much more without that requirement, and this chapter covers those options. As was discussed in
the section on the CAP theorem, it is not possible to keep consistency, availability, and partition tolerance all at the
same time. Thus, by allowing for some divergence between replicas, we can improve the availability of
our systems under a network partition.


What we strive for in this category of replication strategies is **eventual consistency** where nodes may diverge
for sometime, but will be able to sort out the differences once they have contact. We also have 2 options for **eventual consistency**

:Eventual Consistency with probabilistic guarantees:
        Write conflicts are eventually detected and resolved, but it is not always true 
        that the newest value will be used. AWS Dynamo is an example of this type of system.

:Eventual Consistency with strong guarantees:
        Write conflicts are eventually detected and resolved while maintaining a correct
        sequential order of execution. That is, the newest data is always correctly chosen 
        when resolving any given write conflict. All replicas will eventually agree, given 
        that they all receive the same information, regardless of what order the writes were 
        received.

One helper in achieving **Eventual Consistency with strong guarantees** is a **Convergent Replicated
Data Type** (CRDT). A data type is a CRDT if it guarantees convergence to the same value despite the order of
operations. That is, regardless of what order messages come in (and thus operations on the CRDT), the end result will
be the same as long as the set of operations is the same.


Another idea is **Consistency as Logical Monotonicity** (CALM).  **Logically Monotonic** programs are guaranteed to be 
eventually consistent as, by definition, the order of operations is not important.

:Logical Monotonicity:
        If A is a consequence of a set of premises {B}, then A can be inferred by any superset of {B}

What this means for us, is that *new information* cannot invalidated conclusions drawn from previous information. Thus
information (commands) can be provided in any order, and as long as all of the information is *eventually* provided, the 
same results will be reached (*eventual consistency*)

The important qualities for CRDT's are the following:

:Idempotent: Applying the same operation twice using the same operand(s) is the same as applying it only once :math:`A+A=A`.
             An example of this property is adding to a set, if the element is already in the set then set addition of that
             element is a NOP. 

:Associative: This is the property that the order in which operations are applied to a set of operands is not important 
              as long as the **operands** remain in the same order. This is important for a CRDT since the order in which the operands 
              arrive is undetermined. As such it is possible that a replica would apply the operations in a different order without
              re-arranging the operands. For example consider :math:`A+B+C` where '+' is some associative binary operation. Suppose 
              that the replica first receives B and C, it then knows to apply :math:`B+C`. Once the replica receives the operand A it
              would then be able to apply A+(B+C). Since the order of the operands coming in to the replica is undetermined, the order 
              of the operations which will be performed on those operands is also undetermined. Thus associativity is a requirement 
              for a CRDT

:Commutative: Of all the properties this one seems the most obvious. Since the order that any replica will receive the commands is 
              undetermined, it must be able to apply those commands in whichever order it receives them. Else some more complicated 
              communications would have to go on to prevent illegal operations, and the data type would not be a CRDT.
             

The author uses Amazon's Dynamo as a lens to explore the following concepts:


:Consistent Hashing: Using a hash function to map keys to sets of  nodes where they will be stored. This 
                     allows the client to know which subset of all the nodes will have the value for the key 
                     they are looking for. Without consistent hashing, the client would need to poll
                     each group of nodes to see if they had a particular key.

For context, a **quorum** is the minimum number of members of a deliberative assembly necessary to 
conduct the business of that group. In the case of a distributed system, it is the number of nodes 
(votes) required to commit a transaction. 

:Strict Quorums: A **strict quorum system** has the property that any two **quorums** (groups of nodes involved in
                  a voting decision) will always overlap. Requiring a majority of nodes achieves this, as there must have been
                  a node involved in both decisions). Because of this overlapping node, we can be sure that there are no two
                  transactions that result in conflicting data. This is because the vote from that single node which is a part
                  of both quorums is always required to reach a majority. If the transaction being considered is not compatible 
                  with the previous transaction that node will simply not vote in favor, making it impossible to commit an inconsistent 
                  transaction

:Partial Quorums: In a **partial quorum** a majority vote is not required. This means it is possible for there to be two
                  independent quorums which could commit conflicting transactions. Often, the user can choose a number W 
                  (number of nodes required to agree on a WRITE) and a number R (the number of nodes contacted upon a READ).


                  Out of these choices come the following results

                  * R=1, W=N: Fast reads, slow (but consistent) writes
                        * This is total consensus, there will be no conflicts at write time
                  * R=N, W=1: Fast writes, slow reads. Conflicts will be sorted at read time
                  * R=N/2, W=N/2+1: Favorable for both reading and writing. When reading there is a guarantee that
                    one of the nodes contacted would have been involved in any given write transaction, thus the conflicts
                    will be sorted at read time.

                  The property that R + W > N provides strong consistency PROVIDED that the nodes
                  in the set N never change.

                  Systems like Dynamo will swap out nodes when they fail, thus they can only provide strong 
                  consistency probabilistically.


Question/Critique
_____________________

I really enjoyed the chapter and do not have any major criticisms. I was however left wanting to know more 
specifics about CRDTs and what common types of CRDTs are available, as well as library implementations of those
types. 

I recommend having a look at the following sources to get a better handle on CRDTs

1. https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type
3. https://crdt.tech/implementations
2. https://www.microsoft.com/en-us/research/video/strong-eventual-consistency-and-conflict-free-replicated-data-types/?from=http%3A%2F%2Fresearch.microsoft.com%2Fapps%2Fvideo%2Fdl.aspx%3Fid%3D153540 
3. https://www.youtube.com/watch?v=em9zLzM8O7c&t=358s 
4. https://logux.io/ 

This Github page has a great collection of CRDT resouces

https://github.com/alangibson/awesome-crdt






