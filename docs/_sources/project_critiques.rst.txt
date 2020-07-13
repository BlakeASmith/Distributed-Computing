Given what you know of other people's projects in the course (either by the slides posted, and/or hearing about them in lab time), please post three Phase I critiques:
1) one on a project from within your lab section
2) one on a project from outside of your lab section
3) one that is most similar to your own project, which can be within or outside of your lab section! smile

Each critique should be organized as follows:
1) the title of the project, and the person/people involved
2) a summary of project as you understand it (a couple of sentences)
3) the proposed evaluation scenario (a couple of sentences)
4) questions or observations that may help improve the project
5) any references you find on related work that may apply (optional)


Project Phase 1 Critiques
====================================

1) Same Lab Section
-------------------------

For this section I chose the **"Multi-Master Map Reduce"** project from **William Sease**. William
is implementing a simulator of MapReduce with multiple master nodes for improved reliability. In particular, he is making
a simulator (using the Unity3D game engine) which will allow a user to try out different configurations and compare 
the performance statistics. He is planning to use a RAFT-like replication scheme in order to maintain 
an "active" master & some number of "backup" masters.
Otherwise, it will work like a standard MapReduce implementation (like we did for the first set of labs). 


Evaluation
________________________________________

William plans to test the simulator by allowing the user to manually break links, crash servers, and 
otherwise tamper with the system. Presumably (it's not clear from the slides provided) the evaluation 
will be based on manual testing using the manipulations provided by the simulator.


Questions/Observations
___________________________________________

I appreciate that William is making a simulator, allowing for many different options, as opposed to
just doing an implementation of the Mulit-Master MapReduce.

One suggestion is to run a suite of automated tests in addition to the manual testing of the simulator. This 
way you can run the same conditions many times to get the average performance, or monotonically change values 
between iterations to see the trend in performance. Since you will already have functions for making manipulations 
like killing nodes and breaking links, it will be quick to make the automated tests.

I would also recommend having a JSON structure witin your C# code which stores all of the relevent statistics 
(inputs to the simulator & performance characteristics) and logging the results at each run of the simulation. 
Then you could load those structures into Python (or a tool of your choosing) and generate figures/graphs which would
greatly aid in the quality of your final report.

Are the backup Masters involved in the Map-Reduce computation? or are they simply there to be available in the case that
the primary Master fails? if they are involved, how do you handle the situation in which a Master dies while the backup master
is occupied with a Map/Reduce job.

Another question, What work will be run in the simulator? Is it a set task such as a wordcount on a fixed input? or perhaps the 
input is generated?. Does the user get the ability to set the input themselves? 

2) Different Lab Section
------------------------------

**Cameron Metcalfe** and **Akhil Tadimeti** are building a distributed marketplace which uses 
blockchain to record transactions. The marketplace targets students and will allow them to sell & trade 
textbooks and other supplies. The marketplace will take the form of a web application implemented using 
Golang and standard HTML pages. It will allow users to post advertisements of supplies that they are selling 
and will store transactions in the blockchain.


Evaluation
-----------------------------

From the provided slides, the proposed evaluation is not so clear. Nonetheless this is what I got out of them.

1. Ensure transactions occur once and only once
2. Ensure that the approach does not use an unreasonable amount of resources 
   (hard to quantify)

Questions/Critiques
------------------------------

The slides mention that Golang, HTML are used primarily. Presumably for this project to work (and to be distributed) some 
code will need to be executed at the client side. Do you plan to use Javascript for this? Perhaps I'm mistaken and there
is another way which you are planning, but from the slides It's not clear how work is being distributed from the main Golang 
service.

I would also suggest including more information about how specifically the project will be tested. You can use Curl to 
send form data to your HTML forms to mimic user actions. If you are running client side Javascript you can use Selenium 
(the python driver for selenium is great) to run browser instances (which can be headless) and send commands to them. 
You could use this to simulate user actions in the browser.

Here are the docs for Selenium (Python) https://selenium-python.readthedocs.io/

You could also check out grpc-web (https://github.com/grpc/grpc-web) which is a Javascript Grpc client which would
allow you to use RPCs to communicate between your web clients and the Go Server. A standard HTTP API would do you fine 
as well, but gRPC might be something fun to play with for the course. 

3) Most Similar To My Project
------------------------------------

**Mek Obchey** is making a distributed web crawler which will create a graph of 
links between websites. Workers communicate with a Master service via gRPC calls.
urls are pushed to the Worker processes via an RPC call. When the worker receives 
a url to crawl it will request it and pull the links from it. Finally it will clean 
those links (expand relative links, fix broken links, etc) and return them to the master
process.


Evaluation
_____________________________________

From the provided slides, Mek plans to validate the following:

1. Worker RPC (at the master) will return an up-to-date list of worker information
2. Crawl RPC correctly accumulates link information of the URL being crawled
2. Domains RPC returns a correct list of domains crawled
3. Overlap RPC returns the correct number of overlap pages

Questions/Critiques
______________________________________

Here are some questions, critiques, and suggestions I have

How are you going to store the graph information? If the master dies temporarily, 
how will it resume operation?

What is the purpose of the Worker RPC? it returns a list of the Info for each worker 
which has contacted the Master, but what is this information used for? Do the worker 
processes intercommunicate? 

From the slides it seems like the Master sends the worker only one url at a time. 
Do you plan to do any batching to reduce the number of messages between master and workers?

I also think you'll need to come up with some more specifics of how you will test the implementation, 
How will you determine correctness? will you benchmark it under different loads with different numbers of 
workers?


I would also suggest forwarding the urls to each worker based on a hash partition. This 
way the same worker will be responsible for contacting the same servers, which will make it 
easier to control rate limiting. At the master you can accumulate urls from each worker
into a single channel. Then take urls one-by-one out of that channel and separate them (by hash) 
into N channels (assuming you have N workers). Finally assign jobs to the workers from their respective 
channel. When the responses are received at the master simply push all of the newly discovered urls into 
the main channel and let the pipeline do it's work!

This article does a good description of a general approach to what you're trying to do here.
https://medium.com/@morefree7/design-a-distributed-web-crawler-f67a8ebb8336


