SELECT m0.valid_day + 3600 * m0.hour AS avtime,
    SUM(m0.sum2_dtd) AS square_diff_sum,
    SUM(m0.N_dtd) AS N_sum,
    SUM(m0.sum_dtd) AS obs_model_diff_sum,
    SUM(-1 * (m0.sum_dtd - m0.sum_ob_td)) AS model_sum,
    SUM(m0.sum_ob_td) AS obs_sum,
    SUM((if(m0.sum_adtd is not null, m0.sum_adtd, 0))) AS abs_sum
FROM surface_sums2.HRRR_OPS_metar_v2_ALL_HRRR AS m0
WHERE 1 = 1
    AND m0.valid_day + 3600 * m0.hour >= 1675580400
    AND m0.valid_day + 3600 * m0.hour <= 1678172400
    AND m0.hour IN(
        0,
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16,
        17,
        18,
        19,
        20,
        21,
        22,
        23
    )
    AND m0.fcst_len = 9
GROUP BY avtime
ORDER BY avtime;