
Exercises:
========================

:Q1: Have you ever encountered a Heisenbug? How did you isolate and fix it

:A: I'm currently working on a text-processing library for recognizing patterns within
    plain text documents and converting them into python objects. Similarly to what 
    we can already do with JSON but for any format. It uses a string template to define a
    'text object' which can then be recognized in text and it's various attributes extracted
    from the text. This allows any arbitrary data to be read as if it were structured data.

    Repeated function composition is used to alter the matching behaviour as the template 
    string is read. The ultimate result is a function which takes a string as input and
    returns an object with attributes matching the placholders in the template string. 
    There is also recursion involved in this process so it can become somewhat difficult 
    to reason about. 

    I ran into a problem in which the same string template would match a peice of text
    most of the time, except when it didn't. This only happened about 1/10 times and it
    was unclear how to repoduce the bug. 

    My first approach was to log the current remaining text along with the next sub-pattern
    being considered. I ran a script to attemt the match a few thousand times and used grep
    to get all the lines at which the match failed.

    I found that in the faliure cases the pattern being considered did not occur until
    much farther along in the template, but was for some reason being attempted early.

    I isolated the bug to one specific function which was responsible for handling the
    `search` modifer which allows a placholder to be searched for in the text as apposed
    to directly matching the beginning. However there was no indication of what was
    going wrong, and that routine worked in every case I tried to put it through

    Ultimiatley I found that the bug had nothing to do with the routine itself but with
    how the arguments were supplied. Within the routine a wrapper function was defined to 
    add the searching functionallity. The current regex pattern being considered was an
    argument to the enclosing function, and was available to the wrapper function through
    the enclosing scope (closure). It turns out that python uses late-binding, meaning that
    the wrapper function was accessing the value as it was when the wrapper function 
    was **invoked**, where as it needed to access the value as it was at the time when the 
    wrapper function was **defined**. Thus the problem came from the value of that pattern
    variable changing between the time that the function was defined and the time that the
    function was invoked

    Placeholders can be marked as `optional`. If this is the case then the template string
    is evaluated both with and without the placeholder (concurrently). 
    The template string in question
    that was having the inconsistant behavior had several of these optional placeholders.
    It turned out that in most cases, the late-binding did not cause a problem because in 
    the order in which these branches happened to return that specific wrapper 
    function was the last one to be composed with the output function, 
    thus the value of the pattern variable did not change. However some of the time the 
    branches would return in a different order, resulting in the bug manifesting

    The solution was to add `pattern=pattern` in the arguments to the wrapper function.

:Q2: What is the difference between caching and data replication

:A: cached data has an expiration date and is not syncronized to the main dataset.
    typically cached data will only be updated when it is requested. In contrast
    data replication aims to keep the replicated data in the same state as the 
    data which is being replicated. If cached data is continually updated to match
    the service data, then that is data replication and not caching. Caching and data
    replication also differ in purpose. Caching is something you do to improve speed when
    the same data is requested multiple times (you trade accuracy of the data for speed).
    In contrast data replication aims to improve accessabiltiy/availability of the data 
    (in the case that another node fails) and can create overhead (keeping the data in 
    sync). Caching has the effect of reducing the number of requests sent over the network 
    as nodes will use the local cache in place of making a request to another node if 
    possible. Data replication may have effect of increasing the number of requests in the
    network as additional packets must be transmitted to manage the syncronization 
    of the data.
