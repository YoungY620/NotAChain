import pprint
import gantt

def do_plot(filename):
    with open(f"{filename}.log", "r") as f:
        lines = f.readlines()

    time_points = {}

    start_time = 0
    for line in lines:
        if "performance statistic:" in line:
            [key, val] = line.split("performance statistic:")[1].strip().split(": ")
        # print(key, val)
            time_points[key] = int(val)
            if key == "consensus[s][1]":
                start_time = time_points[key]
    # pprint.pprint(time_points)

    for key, val in time_points.items():
        time_points[key] = val - start_time

    bar_map = {}
    for i in range(1, 64):
        for stage in ["consensus", "exe", "commit"]:
            skey, ekey = f"{stage}[s][{i}]", f"{stage}[e][{i}]"
            bar_map[f"{stage}[{i}]"] = (time_points[skey], time_points[ekey]-time_points[skey])

    # pprint.pprint(bar_map)

    bar_map_per_block = {}

    for i in range(1, 64):
        for stage in ["consensus", "exe", "commit"]:
            skey, ekey = f"{stage}[s][{i}]", f"{stage}[e][{i}]"
            if i not in bar_map_per_block:
                bar_map_per_block[i] = []
            bar_map_per_block[i].append([(time_points[skey], time_points[ekey]-time_points[skey])])

    pprint.pprint(bar_map_per_block)

    keys = list(bar_map_per_block.keys())[:3]

    bar_map_per_block = {k: bar_map_per_block[k] for k in keys}
    
    fig = gantt.gantt(["consensus", "execution", "commitment"], bar_map_per_block, [0 for _ in range(len(bar_map_per_block.keys()))], title=filename.split("-")[1])
    fig.savefig(f"{filename}.png")
    
    return fig

fig = do_plot("system-baseline")
fig = do_plot("system-neochain")
