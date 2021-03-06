Clock Exercise
================

:Q: How did you calculate your clients average clock drift rate?
----------------------------

After each successful request to the NTP server I divided the 
calculated clock offset by :math:`10 + 10*numfails` where
**numfails** is the number of failed attempts since the last 
successful attempt. Since a request to the server is made every 10 
seconds this gives a good approximation for the drift per second (in microseconds). 
I kept an array of these values and calculated the average at each iteration. 
The final average was :math:`12.032miliseconds`


Picking the Timeout
-------------------

:Question: what timeout did you pick to detect a failed interaction? What 
    happens if the server's response packet arrives after that timeout

I picked **10 seconds** as the timeout. The resulting packet loss rate was :math:`4%`. 
When a timeout occurred the failure was counted and no other stats were calculated

.. code-block:: python

        while True:
                ...
                try:
                        t3 = calc_time()
                        data, address = client.recvfrom( 1024 )
                        t0 = calc_time()
                        succs += 1
                except Exception as ex:
                        losses += 1
                        logger.debug(f'timeout: {losses}')
                        continue
                ...
                stat = {
                        'offset':off,
                        'RTT':rtt,
                        'smoothed_offset':smoff,
                        'smoothed_RTT':smrtt,
                        'drop_rate':round(losses/(succs+losses), 2)*100,
                        'current_system_time':now,
                        'adjusted_system_time':adjusted,
                        'current_drift':drift,
                        'average_drift':sum(drifts)/len(drifts),
                        'average_RTT':sum(rtts)/len(rtts),
                }
                logger.debug(stat)
                time.sleep(10)

Graphing the Clock Drift
--------------------------

I created this histogram of the clock drift per second (in miliseconds) for each
successful interaction with the server. Note that there were 3200 interactions logged

.. raw:: html
        :file: fig1.html

As you can see most of the time the drift was positive, meaning that my machine's
clock was running faster than the NTP server's. The histogram reflects a fairly normal distribution.

.. raw:: html
        :file: fig2.html

The scatter plot shows that the clock drift skews slightly to the right over time, but the width
(range of the values) stays mostly constant. Adding in the plot of the average drift (calculated at 
each interaction) you can get a better picture of the overall trend. The average slowly increases over time
which means that the client machine is getting out of sync with the NTP server at a faster rate.


Python Code for the NTP client
---------------------------

.. literalinclude:: timeclient.py

Python Code for Parsing the Log and Producing the Graphs
---------------------------

.. literalinclude:: parse_log.py

