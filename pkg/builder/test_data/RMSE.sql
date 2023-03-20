select ceil(
        3600 * floor(((m0.valid_day + 3600 * m0.hour) + 3600 / 2) / 3600)
    ) as avtime,
    count(
        distinct ceil(
            3600 * floor((m0.valid_day + 3600 * m0.hour + 1800) / 3600)
        )
    ) as N_times,
    min(
        ceil(
            3600 * floor((m0.valid_day + 3600 * m0.hour + 1800) / 3600)
        )
    ) as min_secs,
    max(
        ceil(
            3600 * floor((m0.valid_day + 3600 * m0.hour + 1800) / 3600)
        )
    ) as max_secs,
    sum(m0.sum2_dt) as square_diff_sum,
    sum(m0.N_dt) as N_sum,
    sum(m0.sum_dt) as obs_model_diff_sum,
    sum(-1 * (m0.sum_dt - m0.sum_ob_t)) as model_sum,
    sum(m0.sum_ob_t) as obs_sum,
    sum((if(m0.sum_adt is not null, m0.sum_adt, 0))) as abs_sum,
    count(m0.sum2_dt) as N0
from surface_sums2.RAP_NCEP_full_metar_v3u_ALL_HRRR as m0
where 1 = 1
    and m0.valid_day + 3600 * m0.hour >= 1676232000
    and m0.valid_day + 3600 * m0.hour <= 1678824000
    and m0.fcst_len = 6
group by avtime
order by avtime
limit 1;