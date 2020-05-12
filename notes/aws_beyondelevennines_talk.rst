Noes from AWS Beyond Eleven Nines Talk
======================================

:see full talk here: https://www.youtube.com/watch?time_continue=582&v=DzRyrvUF-C0&feature=emb_logo

* trillions of objects, needed to scale over the years

Lessons from security
----------------------
:Threat Model: Get really specific about what a threat is 

:Q: Who is the attacker?, what can they do?, what are they trying to do to us?
:A:

        1. Pick a specific risk (eg. account hacking)
        2. What could they do (eg. password guessing)
        3. What are they tring to do (eg. steal accounts to send spam)

:next: move to pros, answer the questions in detail


Implement Mechanisms
        * Ongoing
        * Technical and organizational approaches

Attacker could be failiure, decay, does not need to be someone

Common Threats
----------------------

:problem: Facility Failiure
        * any single facility can fail at any time

:solution: availability zones, hazard analysis
           -> Buildings that won't fail together

:quote: writing is natures way of showing you how sloppy your thinking is

:problem: Hardware Failiure

worry about the worst exposed componants, not the averages

:quote: you are only as good as you are able to measure yourself

:problem: Data corruption


myths:
* TCP checksums got be covered
* CPU won't lie to me
* Data coruption is a hardware problem
* Corruption is always a single bit flip

:problem: Bugs


* code level bugs
* design level bugs

zonal delpoys
`

Best Practices
----------------------

BYO Threat Models
----------------------
