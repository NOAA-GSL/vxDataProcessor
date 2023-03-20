select ceil(3600 * floor(((m0.time) + 3600 / 2) / 3600)) as avtime,
    sum(m0.yy) as hit,
    sum(m0.yn) as fa,
    sum(m0.ny) as miss,
    sum(m0.nn) as cn
from ceiling_sums2.RAP_OPS_ALL_HRRR as m0
where 1 = 1
    and m0.time >= 1676228400
    and m0.time <= 1678820400
    and m0.trsh = 50
    and m0.fcst_len = 6
group by avtime
order by avtime
limit 1;