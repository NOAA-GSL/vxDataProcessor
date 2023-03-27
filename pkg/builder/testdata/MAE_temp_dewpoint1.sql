select ceil(
        3600 * floor((((m0.valid_day) + 3600 * m0.hour) + 3600 / 2) / 3600)
    ) as avtime,
    count(distinct m0.valid_day + 3600 * m0.hour) as N_times,
    sum(m0.sum2_dt) as square_diff_sum,
    sum(m0.N_dt) as N_sum,
    sum(m0.sum_dt) as obs_model_diff_sum,
    sum(-1 * (m0.sum_dt - m0.sum_ob_t)) as model_sum,
    sum(m0.sum_ob_t) as obs_sum,
    sum((if(m0.sum_adt is not null, m0.sum_adt, 0))) as abs_sum,
    count(m0.sum2_dt) as N0
from vgtyp_sums.RAP_NCEP_full as m0
where 1 = 1
    and m0.valid_day + 3600 * m0.hour >= 1676232000
    and m0.valid_day + 3600 * m0.hour <= 1678824000
    and m0.fcst_len = 6
    and m0.vgtyp IN(1)
group by avtime
order by avtime
limit 1;