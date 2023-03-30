select ceil(
        3600 * floor(
            (
                (unix_timestamp(m0.valid_date) + 3600 * m0.valid_hour) + 3600 / 2
            ) / 3600
        )
    ) as avtime,
    avg(m0.wacorr / 100) as stat
from anom_corr2.GFSDC_anomcorr_7 as m0
where 1 = 1
    and unix_timestamp(m0.valid_date) + 3600 * m0.valid_hour >= { { fromSecs } }
    and unix_timestamp(m0.valid_date) + 3600 * m0.valid_hour <= { { toSecs } }
    and m0.variable = 'HGT'
    and m0.fcst_len = 12
    and m0.level IN(500)
group by avtime
order by avtime;