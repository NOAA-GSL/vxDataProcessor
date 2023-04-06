select ceil(900 * floor(((m0.secs) + 900 / 2) / 900)) as avtime,
    sum(pow(o.direct + o.diffuse - m0.dswrf, 2)) as square_diff_sum,
    count((o.direct + o.diffuse - m0.dswrf)) as N_sum,
    sum((o.direct + o.diffuse - m0.dswrf)) as obs_model_diff_sum,
    sum(
        (
            if(o.direct + o.diffuse is not null, m0.dswrf, null)
        )
    ) as model_sum,
    sum(
        (
            if(m0.dswrf is not null, o.direct + o.diffuse, null)
        )
    ) as obs_sum,
    sum((abs(o.direct + o.diffuse - m0.dswrf))) as abs_sum
from surfrad3.surfrad as o,
    surfrad3.RAP_OPS_130 as m0
where 1 = 1
    and m0.id = o.id
    and m0.secs = o.secs
    and o.secs >= 1676257200
    and o.secs <= 1678849200
    and m0.secs >= 1676257200
    and m0.secs <= 1678849200
    and m0.fcst_len = 360
    and m0.scale = 26
group by avtime
order by avtime
limit 1;