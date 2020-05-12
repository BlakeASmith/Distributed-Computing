import re
import pandas
import numpy
import plotly.express as px
import plotly.graph_objects as go
from types import SimpleNamespace

with open('rtt_offset.log') as log:
    data = re.findall('{.*?}', log.read())

stats = [eval(it) for it in data]

avg_RTT = stats[-1]['average_RTT']/1000000000
packet_loss_rate = stats[-1]['drop_rate']
avg_drift_in_milis = stats[-1]['average_drift']/1000000


for stat in stats:
    stat['current_drift'] /= 1000000
    stat['average_drift'] /= 1000000

frame = pandas.DataFrame(stats)

fig = px.histogram(frame, x='current_drift', 
        marginal = 'violin',
        title='Histogram of Clock Drift Per Second', 
        labels={'current_drift':'drift_in_miliseconds_per_second', 'y':'percent of records'}, 
        opacity=0.7, 
        color_discrete_sequence=['indianred'],
        hover_data=frame.columns)
fig.update_layout(xaxis_title="Drift in miliseconds/sec", yaxis_title="Count")
fig.write_html('fig1.html')

fig2 = px.scatter(frame, x='current_drift')
fig2.add_trace(go.Scattergl(
    x=frame.average_drift, 
    mode='markers', 
    name='average drift at each interaction',
     marker=dict(
        size=10,
        color=numpy.random.randn(1000), #set color equal to a variable
        colorscale='Viridis', # one of plotly colorscales
        line_width=1
    )
))
fig2.update_layout(title='Scatterplot of Clock Drift', 
        xaxis_title="Drift in miliseconds/sec", 
        yaxis_title="Count")
fig2.show()
fig2.write_html('fig2.html')



