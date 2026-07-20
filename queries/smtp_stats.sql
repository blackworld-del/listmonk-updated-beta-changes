-- smtp_stats

-- name: get-smtp-profile-stats
SELECT
    COALESCE(SUM(sl.sent), 0) AS total_sent,
    COALESCE(SUM(sl.failed), 0) AS total_failed,
    CASE WHEN COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0) > 0
        THEN ROUND(COALESCE(SUM(sl.sent), 0)::NUMERIC / (COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0)) * 100, 1)
        ELSE 0
    END AS success_rate,
    COALESCE(SUM(sl.sent) FILTER (WHERE sl.created_at >= CURRENT_DATE), 0) AS sent_today,
    COALESCE(SUM(sl.failed) FILTER (WHERE sl.created_at >= CURRENT_DATE), 0) AS failed_today,
    COALESCE(SUM(sl.sent) FILTER (WHERE sl.created_at >= date_trunc('week', CURRENT_DATE)), 0) AS sent_week,
    COALESCE(SUM(sl.sent) FILTER (WHERE sl.created_at >= date_trunc('month', CURRENT_DATE)), 0) AS sent_month,
    MAX(sl.end_time) AS last_sent_at,
    AVG(EXTRACT(EPOCH FROM (sl.end_time - sl.start_time))) AS avg_send_time_seconds,
    (SELECT campaign_id FROM smtp_logs WHERE smtp_profile_id = $1 ORDER BY end_time DESC NULLS LAST LIMIT 1) AS last_campaign_id
FROM smtp_logs sl
WHERE sl.smtp_profile_id = $1;

-- name: get-smtp-profile-daily-stats
SELECT
    sl.created_at::DATE AS date,
    COALESCE(SUM(sl.sent), 0) AS sent,
    COALESCE(SUM(sl.failed), 0) AS failed,
    CASE WHEN COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0) > 0
        THEN ROUND(COALESCE(SUM(sl.sent), 0)::NUMERIC / (COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0)) * 100, 1)
        ELSE 0
    END AS success_rate
FROM smtp_logs sl
WHERE sl.smtp_profile_id = $1
    AND sl.created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY sl.created_at::DATE
ORDER BY sl.created_at::DATE ASC;

-- name: get-smtp-profile-recent-campaigns
SELECT
    c.id,
    c.name,
    c.started_at,
    c.updated_at AS end_time,
    c.sent,
    c.to_send,
    c.status,
    COALESCE(sl.failed, 0) AS failed
FROM campaigns c
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(failed), 0) AS failed
    FROM smtp_logs
    WHERE smtp_profile_id = $1 AND campaign_id = c.id
) sl ON true
WHERE c.smtp_profile_id = $1
ORDER BY c.updated_at DESC
LIMIT $2
OFFSET $3;

-- name: get-smtp-profile-activity
SELECT id, smtp_profile_id, event_type, message, created_at
FROM smtp_activity_log
WHERE smtp_profile_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: insert-smtp-log
INSERT INTO smtp_logs (smtp_profile_id, campaign_id, sent, failed, start_time, end_time)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: insert-smtp-activity
INSERT INTO smtp_activity_log (smtp_profile_id, event_type, message)
VALUES ($1, $2, $3)
RETURNING id;

-- name: get-smtp-overview
SELECT
    sp.id,
    sp.name,
    sp.host,
    sp.username,
    sp.enabled,
    COALESCE(s.sent_today, 0) AS sent_today,
    COALESCE(s.total_sent, 0) AS total_sent,
    COALESCE(s.total_failed, 0) AS total_failed,
    COALESCE(s.success_rate, 0) AS success_rate,
    s.last_sent_at
FROM smtp_profiles sp
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM(sl.sent), 0) AS total_sent,
        COALESCE(SUM(sl.failed), 0) AS total_failed,
        CASE WHEN COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0) > 0
            THEN ROUND(COALESCE(SUM(sl.sent), 0)::NUMERIC / (COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0)) * 100, 1)
            ELSE 0
        END AS success_rate,
        COALESCE(SUM(sl.sent) FILTER (WHERE sl.created_at >= CURRENT_DATE), 0) AS sent_today,
        MAX(sl.end_time) AS last_sent_at
    FROM smtp_logs sl
    WHERE sl.smtp_profile_id = sp.id
) s ON true
ORDER BY sp.name ASC;

-- name: get-smtp-dashboard-stats
SELECT
    COUNT(*) AS total_profiles,
    COUNT(*) FILTER (WHERE enabled = true) AS active_profiles,
    COUNT(*) FILTER (WHERE enabled = false) AS disabled_profiles,
    COALESCE(SUM(sl.sent) FILTER (WHERE sl.created_at >= CURRENT_DATE), 0) AS emails_sent_today,
    COALESCE(SUM(sl.failed) FILTER (WHERE sl.created_at >= CURRENT_DATE), 0) AS failed_today
FROM smtp_profiles sp
LEFT JOIN smtp_logs sl ON sl.smtp_profile_id = sp.id;

-- name: export-smtp-stats
SELECT
    sl.created_at::DATE AS date,
    sp.name AS smtp_profile,
    COALESCE(SUM(sl.sent), 0) AS emails_sent,
    COALESCE(SUM(sl.failed), 0) AS failed_emails,
    CASE WHEN COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0) > 0
        THEN ROUND(COALESCE(SUM(sl.sent), 0)::NUMERIC / (COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0)) * 100, 1)
        ELSE 0
    END AS success_rate
FROM smtp_logs sl
JOIN smtp_profiles sp ON sp.id = sl.smtp_profile_id
GROUP BY sl.created_at::DATE, sp.name
ORDER BY sl.created_at::DATE DESC, sp.name ASC;

-- name: query-smtp-profiles-stats
SELECT sp.*,
    COALESCE(c.campaign_count, 0) AS campaign_count,
    COALESCE(s.sent_today, 0) AS sent_today,
    COALESCE(s.total_sent, 0) AS total_sent,
    COALESCE(s.total_failed, 0) AS total_failed,
    COALESCE(s.success_rate, 0) AS success_rate,
    s.last_sent_at
FROM smtp_profiles sp
LEFT JOIN (
    SELECT smtp_profile_id, COUNT(*) AS campaign_count
    FROM campaigns
    WHERE smtp_profile_id IS NOT NULL
    GROUP BY smtp_profile_id
) c ON c.smtp_profile_id = sp.id
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM(sl.sent), 0) AS total_sent,
        COALESCE(SUM(sl.failed), 0) AS total_failed,
        CASE WHEN COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0) > 0
            THEN ROUND(COALESCE(SUM(sl.sent), 0)::NUMERIC / (COALESCE(SUM(sl.sent), 0) + COALESCE(SUM(sl.failed), 0)) * 100, 1)
            ELSE 0
        END AS success_rate,
        COALESCE(SUM(sl.sent) FILTER (WHERE sl.created_at >= CURRENT_DATE), 0) AS sent_today,
        MAX(sl.end_time) AS last_sent_at
    FROM smtp_logs sl
    WHERE sl.smtp_profile_id = sp.id
) s ON true
ORDER BY %order%;
