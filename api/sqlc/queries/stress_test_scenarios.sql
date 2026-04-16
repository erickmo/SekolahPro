-- Stress Test Scenario queries.

-- name: GetStressTestByID :one
SELECT id, scenario_name, scenario_type, description, parameters,
       projected_car, projected_npl, projected_roa, projected_roe,
       capital_adequate, survives_scenario, test_date, created_at, created_by, deleted_at
FROM stress_test_scenarios
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListStressTests :many
SELECT id, scenario_name, scenario_type, description, parameters,
       projected_car, projected_npl, projected_roa, projected_roe,
       capital_adequate, survives_scenario, test_date, created_at, created_by
FROM stress_test_scenarios
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountStressTests :one
SELECT COUNT(*)
FROM stress_test_scenarios
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertStressTest :one
INSERT INTO stress_test_scenarios (
    id, tenant_id, company_id, scenario_name, scenario_type, description,
    parameters, projected_car, projected_npl, projected_roa, projected_roe,
    capital_adequate, survives_scenario, test_date, created_at, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING id;

-- name: UpdateStressTestResults :exec
UPDATE stress_test_scenarios
SET projected_car = $4, projected_npl = $5, projected_roa = $6, projected_roe = $7,
    capital_adequate = $8, survives_scenario = $9
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteStressTest :exec
UPDATE stress_test_scenarios
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
