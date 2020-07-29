Event Streaming: Lessons from Confluent-Cloud
=====================================================

See the video here: https://www.youtube.com/watch?time_continue=2&v=YvVf97xeYkw

In this Keynote Neha Narkhede, the co-founder and CTO at Confluent, goes over the pros of   
**event streaming** via Apache Kafka and more specifically the services offered through *Confuent Cloud*. 
Event streaming (treating data as a continuous stream of events) provides persistent, ordered, and scalable 
storage that an ETL (Extract, Transform, Load) system would while also providing low latency. This bridges the 
gap between traditional databases, which deal with persistence, and messaging services, which deal with communication 
between services. 

In her talk, Narkhede describes ETL as considering "what happened in the world" and messaging services as "what is happenning in the world".
In contrast, she describes Kafka as "what is **contextually** happening in the world". In addition to the benefits of **event streaming** as 
a paradigm for data storage, Narkhede discusses the features and benefits of the services which are offered by Confuent, build on top of Apache Kafka 
and other cloud services.


Such services include KSQL, the KSQL processing log, Confluent-Cloud, and Confluent-Platform.
KSQL is a query language built for use with Kafka which provides an SQL-like experience for reading the 
Kafka event stream. ksqlDB includes a processing log which stores any errors and details how each row is 
processed. The interesting thing about the processing log is that it is itself an event stream which can be 
queried using KSQL. This allows access to the logs via a structured schema, easing debugging of KSQL apps. 


**Confluent Cloud** is a fully managed event streaming platform which is based on Kafka. It runs on infrastructure provided
by Google Cloud, AWS, Microsoft Azure, etc. but is managed by Confluent. This means that Confluent has already went through the 
trouble of optimizing Apache Kafka to scale well over the cloud infrastructure provided by Amazon, Google, Microsoft, etc. 
Because of this, Confluent-Cloud offers seamless scaling from 0mb/s to 100mb/s and back without having to deal with any cluster sizing. 
It is also a pay for what you use service, which makes it reasonably priced for small projects. Confluent-Cloud aims to get the 
developer up and running with Kafka within minutes and handle the complexity of cluser-sizing and configuration on their behalf.

**Confluent Platform** is a locally managed version of Confluent-Cloud which allows developers to handle their own deployment 
while still utilizing the features of KSQL and other perks provided by Confluent.


In my own projects I have found streams (& lazily evaluated sequences) to be an invaluable tool, especially in regards to 
concurrent & asynchronous programming. Through this experience I can see how it would be useful to model data this way at 
a larger scale. I also think that a streaming model fits best with the way that things actually are in distributed systems
(which are naturally asynchronous).

Though not particularly a criticism of the talk, I do find myself wanting to know more about what Kafka is and how it works.
It is difficult to appreciate the benefits of Confluent's services without first understanding Kafka and the difficulties surrounding it 

I did some researching on Kafka from these sources:

* The official documentation 

  * https://kafka.apache.org/documentation/#introduction
* This Kafka producer tutorial for *Kotlin* 

  * https://aseigneurin.github.io/2018/08/01/kafka-tutorial-1-simple-producer-in-kotlin.html
* This producer-consumer tutorial for *Kotlin* 

  * https://dzone.com/articles/producer-consumer-with-kafka-and-kotlin
* This simple (entry-level) guide to Kafka in Python 

  * https://towardsdatascience.com/kafka-python-explained-in-10-lines-of-code-800e3e07dad1
* This word-count example in Java using Kafka streams api 

  * https://github.com/confluentinc/kafka-streams-examples/blob/5.5.0-post/src/main/java/io/confluent/examples/streams/WordCountLambdaExample.java
* This quickstart for running Kafka using Confluent Platform and Docker 

  * https://docs.confluent.io/current/quickstart/ce-docker-quickstart.html


I wrote a simple wordcount program which uses Kafka as an example, you can see the repo at https://github.com/BlakeASmith/KafkaExample


