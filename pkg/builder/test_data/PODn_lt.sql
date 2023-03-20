select ceil(3600 * floor(((m0.time) + 3600 / 2) / 3600)) as avtime,
    sum(m0.yy) as hit,
    sum(m0.ny) as fa,
    sum(m0.yn) as miss,
    sum(m0.nn) as cn
from cref.RAP_OPER_13km_CONUS as m0
where 1 = 1
    and m0.time >= 1676120400
    and m0.time <= 1678712400
    and m0.trsh = 15
    and m0.fcst_len = 6
group by avtime
order by avtime
limit 1;