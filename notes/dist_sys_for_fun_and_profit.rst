Distributed Systems for Fun and Profit
===========================================

:Question from prelab exercise:
        What do you find most interesting about Chapter 1 
        (Distributed systems at a high level), and why?

:My Answer:

        I found that Takada gave interesting explanations for many of the trade-offs 
        involved in distributed computing. 
        (Performance vs Availability, Intelligibility vs Performance, etc). 


        What was most interesting to me, from the first chapter, was the notion
        that oftentimes availability must be sacrificed in order 
        to maintain guarantees about latency. Thus 
        exposing the trade-off between availability & latency guarantees. 

        I also enjoyed the explanation at the beginning of the chapter 
        regarding the theoretical lack of need for distributed systems in 
        the first place. Takada expresses the notion that, given infinite r&d 
        resources there would be no requirement, or benefit, to splitting a 
        computation across multiple nodes. Instead the hardware of a single machine
        could be indefinitely upgraded. 

        The reason that distributed systems are required in practice is only due
        to the limitations of our current technology.  

Two consequences of distribution:

* Information travels at the speed of light
* **independent things fail independently**

Notes from Chapter 1
---------------------------------------

As system requirements for a computation increase, a point 
is reached at which it is no longer feasible (or is cost prohibitive)
to upgrade a single machine to do the job. In consequence of this 
it becomes a requirement to run computations over multiple machines 
(a distributed system).

:Definition of **Scalability**: 
        The ability of a system to be enlarged to accommodate
        a growing amount of work

In other words, as the size of the problem increases the performance 
should get **incrementally worse**.

:Size Scalability: adding more nodes to the problem should make the 
                   system **linearly** faster. Ideally 3x as many nodes
                   would mean 3x as fast performance, or 3x as much data
                   processed in the same amount of time.

:Geographic Scalability: possibility of adding more data centres in different 
                         locations to increase response times (dealing with 
                         cross-data centre latency reasonably).

:Admin Scalability: The ability to add nodes to the system without 
                    increasing administrative costs.


Performance & Availability 
_______________________________________

**Performance** is the amount of work accomplished by a system
in comparison to the resources (including time) used.

In contrast, **Availability** is the *proportion* of time in which 
the system functions as intended. 

.. math::

        Availability = \frac{uptime}{uptime + downtime}


Latency vs Latent
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

:**Latency**:  The period of delay between the initiation of something
               and the occurrence of it.

               - example; how long does it take for a write to be visible 
                 to readers?

:Latent: Something is **latent** if it exists but is inactive.

Availability & Fault Tolerance
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Fault Tolerance** is the ability of a system to
continue to behave (in a well-defined way) when faults occur.

.. note::

        You cannot tolerate faults which you have not considered

        ==> It is important to consider what can go wrong in a distributed 
        system and design the system appropriatley 

What's in the Way?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

A distributed system is constrained by both the **number of nodes**
and the **distance between them**.

Thus the following is clear:

* more nodes => higher probability of failure
* more nodes => more communication overhead
* farther apart nodes => increased latency of communication

I use the notation '=>' above since these factors are 
universally true (tautology) as they are a consequence of 
**physical constraints**. 

Any other factors are a consequence of a system's **design**.


Intelligibility vs Efficiency 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Generally speaking, there is a trade off between 
how easy a design is to understand and it's performance. 

A system which makes **weak guarentees** has more freedom of action, 
and thus more potential for performance. However it also becomes 
more complicated to reason about.

When failures occur which increase the latency of a system, a system 
must decide whether do reduce availability, & thus maintain it's 
guarantees about performance, or to stay available and fail to meet
it's performance guarantees.

In this sense, there is a trade off between **availabilty** and **Performance**.

Technique: Partition & Replicate
_______________________________________

A Divide & Conquer approach

1. Partition any data that can be partitioned and provide one 
   partition to each node
2. Replicate any data which cannot be partitioned and provide 
   a copy of that data to each node

:A problem with replication:
        maintaining multiple copies of the same data requires some way to
        keep that data synchronised and/or deal with inconsistencies.

Furthermore, there is a trade off between the strength of consistency 
and latency. A weaker consistency model will be faster than a stronger one.

















