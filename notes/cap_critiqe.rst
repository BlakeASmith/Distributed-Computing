Summary of A Critique of the CAP Theorem 
============================================

See the full paper at https://arxiv.org/pdf/1509.05393.pdf

This paper discusses some problems regarding the CAP theorem including 
the unclear distinction between *strong* and *eventual* consistency. It also 
presents some alternatives to CAP like a *delay sensitivity* framework.

Some of the authors gripes with the CAP theorem include:

:No Maximum Latency: CAP's definition of availability does not 
                     include an upper bound for latency. A system
                     is considered available as long as it *eventually*
                     responds. This is not consistent with the real world 
                     when there really is an upper bound on acceptable 
                     response times.

:Spectrum of Consistency: The CAP theorem only distinguishes between *strong* (single-copy) consistency 
                          and *weak* (eventual) consistency. In reality there are many forms of consistency
                          appropriate for different applications. These many forms of consistency are not
                          properly considered in CAP's definitions (CAP is overly simplistic). Additionally 
                          the "Strong" consistency in CAP is sometimes defined as **linearalizability**, and 
                          other times defined as **one-copy serialiazabiltiy**. By simply stating the CAP theorem 
                          it is not clear what is meant by "Strong consistency".

:Partition Tolerance is defined by the System Model: 
        It is not quite accurate to state that an algorithm provides *partition tolerance*, since what 
        "partition tolerance" means depends on the system model (set of assumptions about the system and network).
        For example an algorithm could be tolerant to short periods of unavailability, but not long periods. 

:Theorem 1 only holds where Infinite length partitions are possible:
        Thm 1 states that it is impossible for an asynchronous network 
        to provide *availaiblility* and *atomic consistency* (for a read/write object).
        However, in the case of a network with **fair-loss links** (there is a non-zero
        probability of a message transmitting successfully) then these properties can be 
        trivially satisfied by simply retrying messages until they arrive. Since having fair-loss 
        links guarantees delivery in an (unbounded but) finite amount of time, CAP's definition
        of availability will still be satisfied. 

:System Model assumed by Thm 1 makes eventual consistency impossible:
        For Thm 1 to hold it must be possible for partitions of infinite 
        duration to occur. Thm 1 correctly states that availability and 
        strong consistency (linearalizability) cannot both hold under these 
        conditions. However, eventual consistency is also impossible under these
        conditions (since it is possible that two nodes will **never again** be able
        to communicate).



The paper goes on to suggest that an eventual response is too weak a
requirement for availability, and that availability should be considered in
terms of latency. They propose a model of **delay-sensitivity** as an alternative.





Questions/Criticisms
---------------------------

I appreciate that the paper provided a viable alternative to the CAP theorem 
while criticising it. However I did feel that some of the criticism given, though
well justified, misinterpreted the purpose of the CAP theorem.

For example, the authors criticize CAP for it's definition of availability not 
requiring an upper bound. Though in real world application an upper bound is strictly 
necessary, It still may be useful to consider algorithms which provide which do not 
have bounded response times. CAP is concerned with what is theoretically possible, not 
necessarily what is realistically possible. It is fair to use an alternative model which
better reflects the real-world requirements, but it does not invalidate CAP in theory.


Related Resources
----------------------------

This article (https://martin.kleppmann.com/2015/05/11/please-stop-calling-databases-cp-or-ap.html) hits on many 
of the same points as the paper, while being a little easier to digest.

This post (http://www.shantala.io/cap-theorem/) gives a simple explanation of the consistency/latency trade-off
discussed in the paper







