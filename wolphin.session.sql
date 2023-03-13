select ceil(
        3600 * floor(
            ((unix_timestamp(m0.date) + 3600 * m0.hour) + 3600 / 2) / 3600
        )
    ) as avtime,
    count(distinct unix_timestamp(m0.date) + 3600 * m0.hour) as N_times,
    min(unix_timestamp(m0.date) + 3600 * m0.hour) as min_secs,
    max(unix_timestamp(m0.date) + 3600 * m0.hour) as max_secs,
    sum(m0.sum2_dt) as square_diff_sum,
    sum(m0.N_dt) as N_sum,
    sum(m0.sum_dt) as obs_model_diff_sum,
    sum(-1 * (m0.sum_dt - m1.sum_ob_t)) as model_sum,
    sum(m1.sum_ob_t) as obs_sum,
    sum(0) as abs_sum,
    group_concat(
        unix_timestamp(m0.date) + 3600 * m0.hour,
        ';',
        m0.mb10 * 10,
        ';',
        m0.sum2_dt,
        ';',
        m0.N_dt,
        ';',
        m0.sum_dt,
        ';',
        -1 * (m0.sum_dt - m1.sum_ob_t),
        ';',
        m1.sum_ob_t,
        ';',
        0
        order by unix_timestamp(m0.date) + 3600 * m0.hour,
            m0.mb10 * 10
    ) as sub_data,
    count(m0.sum2_dt) as N0
from ruc_ua_sums2.RAP_OPS_iso_Areg14 as m0,
    ruc_ua_sums2.HRRR_OPS_Areg14 as m1
where 1 = 1
    and m0.date = m1.date
    and m0.hour = m1.hour
    and m0.mb10 = m1.mb10
    and unix_timestamp(m0.date) + 3600 * m0.hour >= 1675450800 - 1800
    and unix_timestamp(m0.date) + 3600 * m0.hour <= 1678042800 + 1800
    and m0.fcst_len = 6
    and m1.fcst_len = 0
    and m0.mb10 >= 1 / 10
    and m0.mb10 <= 1000 / 10
group by avtime
order by avtime;