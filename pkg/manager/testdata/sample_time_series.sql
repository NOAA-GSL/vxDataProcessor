SELECT m0.valid_day + 3600 * m0.hour AS avtime,
    SUM(m0.sum2_dw) AS square_diff_sum,
    SUM(m0.N_dw) AS N_sum,
    SUM(m0.sum_ob_ws - m0.sum_model_ws) AS obs_model_diff_sum,
    SUM(m0.sum_model_ws) AS model_sum,
    SUM(m0.sum_ob_ws) AS obs_sum,
    SUM(0) AS abs_sum
FROM surface_sums2.HRRR_OPS_metar_v2_ALL_HRRR AS m0
WHERE 1 = 1
    AND m0.valid_day + 3600 * m0.hour >= 1675209600
    AND m0.valid_day + 3600 * m0.hour <= 1677628800
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
    AND m0.fcst_len = 0
GROUP BY avtime
ORDER BY avtime;
