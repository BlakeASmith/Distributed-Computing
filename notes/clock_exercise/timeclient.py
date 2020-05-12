import socket
import struct
import sys
import time
import logging
import math
NTP_SERVER = "0.uk.pool.ntp.org"
TIME1970 = 2208988800

logger = logging.getLogger('rtt_and_offset')
logger.setLevel(logging.DEBUG)
handler = logging.FileHandler('rtt_offset.log')
handler.setLevel(logging.DEBUG)
logger.addHandler(handler)


SOCKET_TIMEOUT = 10
NANOS = 1000000000

def calc_time():
    t = time.time()
    return (t//1, (t%1)*NANOS//1)

def diff(a, b):
    secs = b[0]-a[0]
    nanos = b[1]-a[1]
    nanos = (NANOS)*secs + nanos

    return nanos

def addnanos(it, nanos):
    newnanos, secs = it[1]+nanos, it[0]
    if newnanos >= NANOS:
        secs += 1 
        newnanos -= NANOS
    return (secs, newnanos)

stats, drifts, rtts = [], [], []
succs, losses, cur_fails = 0, 0, 0
while True:
    data = '\x1b' + 47 * '\0'
    data = data.encode('utf-8')
    client = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    client.settimeout(SOCKET_TIMEOUT)
    client.sendto( data,( NTP_SERVER, 123 ))
    try:
        t3 = calc_time()
        data, address = client.recvfrom( 1024 )
        t0 = calc_time()
        succs += 1
    except Exception as ex:
        losses += 1
        logger.debug(f'timeout: {losses}')
        continue

    resp = struct.unpack( '!12I', data )

    reference = (resp[4]-TIME1970, resp[5]%NANOS)
    originate = (resp[6]-TIME1970, resp[7]%NANOS)
    receive   = (resp[8]-TIME1970, resp[9]%NANOS)
    transmit  = (resp[10]-TIME1970, resp[11]%NANOS)

    t1, t2 = transmit, receive

    off, rtt = (diff(t3,t2) - diff(t1,t0))/2, diff(t3, t0)
    drift = off/((10+(rtt/NANOS))*(losses-cur_fails+1))
    drifts.append(drift)
    cur_fails = losses

    rtts.append(rtt)

    stats.append((rtt, off))
    if len(stats) >= 8:
        stats = stats[1:]

    smrtt, smoff = min(stats, key=lambda st: st[0])

    now = calc_time()
    adjusted = addnanos(now, smoff)

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



