WITH average_times AS (
    SELECT
    user_id,
    CASE 
        WHEN EXTRACT(DOW FROM date) = 6 THEN (time_in_seconds / 49.0) * 25.0
        ELSE time_in_seconds
    END AS normalized_time
    FROM times
), user_percentiles AS (
    SELECT 
    user_id,
    PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY normalized_time) AS percentile_10,
    PERCENTILE_CONT(0.8) WITHIN GROUP (ORDER BY normalized_time) AS percentile_20,
    PERCENTILE_CONT(0.7) WITHIN GROUP (ORDER BY normalized_time) AS percentile_30,
    PERCENTILE_CONT(0.6) WITHIN GROUP (ORDER BY normalized_time) AS percentile_40,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY normalized_time) AS percentile_50,
    PERCENTILE_CONT(0.4) WITHIN GROUP (ORDER BY normalized_time) AS percentile_60,
    PERCENTILE_CONT(0.3) WITHIN GROUP (ORDER BY normalized_time) AS percentile_70,
    PERCENTILE_CONT(0.2) WITHIN GROUP (ORDER BY normalized_time) AS percentile_80,
    PERCENTILE_CONT(0.1) WITHIN GROUP (ORDER BY normalized_time) AS percentile_90
    FROM average_times
    GROUP BY user_id
), all_percentiles AS (
    SELECT 
    -1 AS user_id,
    'Everyone' AS name,
    PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY normalized_time) AS percentile_10,
    PERCENTILE_CONT(0.8) WITHIN GROUP (ORDER BY normalized_time) AS percentile_20,
    PERCENTILE_CONT(0.7) WITHIN GROUP (ORDER BY normalized_time) AS percentile_30,
    PERCENTILE_CONT(0.6) WITHIN GROUP (ORDER BY normalized_time) AS percentile_40,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY normalized_time) AS percentile_50,
    PERCENTILE_CONT(0.4) WITHIN GROUP (ORDER BY normalized_time) AS percentile_60,
    PERCENTILE_CONT(0.3) WITHIN GROUP (ORDER BY normalized_time) AS percentile_70,
    PERCENTILE_CONT(0.2) WITHIN GROUP (ORDER BY normalized_time) AS percentile_80,
    PERCENTILE_CONT(0.1) WITHIN GROUP (ORDER BY normalized_time) AS percentile_90
    FROM average_times
)
SELECT users.id, users.name, user_percentiles.percentile_10, user_percentiles.percentile_20, user_percentiles.percentile_30, user_percentiles.percentile_40, user_percentiles.percentile_50, user_percentiles.percentile_60, user_percentiles.percentile_70, user_percentiles.percentile_80, user_percentiles.percentile_90
FROM user_percentiles
JOIN users ON user_percentiles.user_id = users.id
UNION ALL
SELECT * FROM all_percentiles
ORDER BY id ASC;