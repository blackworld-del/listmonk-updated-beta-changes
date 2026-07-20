-- smtp_profiles

-- name: query-smtp-profiles
SELECT sp.*, COALESCE(c.campaign_count, 0) AS campaign_count
FROM smtp_profiles sp
LEFT JOIN (
    SELECT smtp_profile_id, COUNT(*) AS campaign_count
    FROM campaigns
    WHERE smtp_profile_id IS NOT NULL
    GROUP BY smtp_profile_id
) c ON c.smtp_profile_id = sp.id
ORDER BY sp.name ASC;

-- name: get-smtp-profile
SELECT sp.*, COALESCE(c.campaign_count, 0) AS campaign_count
FROM smtp_profiles sp
LEFT JOIN (
    SELECT smtp_profile_id, COUNT(*) AS campaign_count
    FROM campaigns
    WHERE smtp_profile_id IS NOT NULL
    GROUP BY smtp_profile_id
) c ON c.smtp_profile_id = sp.id
WHERE sp.id = $1 OR sp.uuid = $2;

-- name: create-smtp-profile
INSERT INTO smtp_profiles (uuid, name, host, port, username, password, encryption, from_email, from_name, reply_to, enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: update-smtp-profile
UPDATE smtp_profiles SET
    name = $2,
    host = $3,
    port = $4,
    username = $5,
    password = COALESCE(NULLIF($6, ''), password),
    encryption = $7,
    from_email = $8,
    from_name = $9,
    reply_to = $10,
    enabled = $11,
    updated_at = NOW()
WHERE id = $1;

-- name: delete-smtp-profile
DELETE FROM smtp_profiles WHERE id = $1;

-- name: get-smtp-profile-by-name
SELECT id FROM smtp_profiles WHERE name = $1;
