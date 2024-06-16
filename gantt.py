import enum
import matplotlib.pyplot as plt
import matplotlib.patches as mpatch
import numpy as np


def gantt(category_names, results, arrival_t, title="Gantt Chart"):
    fig, gnt = plt.subplots(figsize=(10, 3))

    lables = list(results.keys())
    num_lables = len(lables)
    data = list(results.values())
    max_time = max([sum(seg)
                    for lable in data for segs in lable for seg in segs])

    yticks = np.linspace(num_lables*10+5, 15, num_lables)

    gnt.set_ylim(0, num_lables*10+20)
    gnt.set_xlim(0, max_time*1.06)

    gnt.set_xlabel('seconds since start')
    gnt.set_ylabel('Process')

    gnt.set_yticks(yticks)
    gnt.set_yticklabels(lables)

    # category_colors = plt.get_cmap('Greys')(
    #     np.linspace(0.15, 0.85, len(category_names)))
    # category_colors = plt.get_cmap('tab20')(np.linspace(0, 1, len(category_names)))
    category_colors = plt.get_cmap()(np.linspace(0, 1, len(category_names)))


    gnt.grid(False)

    seg_height = 6

    for i, data in enumerate(data):
        y_lb = yticks[i]-3
        y_ub = y_lb+seg_height
        for (segs, color) in zip(data, category_colors):
            gnt.broken_barh(segs, (y_lb, seg_height), fc=color)
            gnt.errorbar(arrival_t[i], yticks[i], seg_height/2)
            if arrival_t[i] != 0:
                gnt.text(arrival_t[i], y_ub+1, str(arrival_t[i]),
                         ha='center', va='bottom', color='blue')

    fakebar = [mpatch.Rectangle((0, 0), 1, 1, fc=category_color)
               for category_color in category_colors]

    gnt.legend(fakebar, category_names)
    gnt.set_title(title)
    fig.tight_layout() 
    
    return plt