package vernon

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// BaseService menyediakan operasi bisnis generik untuk semua domain Vernon.
// Write flow: validate → loadAutoloadData → INSERT/UPDATE dalam TX → COMMIT → emitSync.
type BaseService struct {
	db   *sqlx.DB
	repo *BaseRepository
	desc DomainDescriptor
	eb   eventbus.EventBus
	log  zerolog.Logger
}

// NewBaseService membuat instance baru BaseService.
func NewBaseService(db *sqlx.DB, repo *BaseRepository, desc DomainDescriptor, eb eventbus.EventBus, log zerolog.Logger) *BaseService {
	return &BaseService{db: db, repo: repo, desc: desc, eb: eb, log: log}
}

// Create membuat entitas baru. Autoload data di-resolve sinkron di dalam TX.
func (s *BaseService) Create(ctx context.Context, sc scope.Scope, data map[string]any) (*BaseDomain, error) {
	if err := s.desc.Validate(data); err != nil {
		return nil, err
	}
	rels := s.desc.DefaultRels()

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	enriched, err := s.loadAutoloadData(ctx, sc, rels, data)
	if err != nil {
		return nil, err
	}

	entity, err := s.repo.Create(ctx, tx, sc, rels, enriched)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", s.desc.TableName(), err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.emitSync(ctx, sc, entity.ID.String(), ActionCreate, nil)
	return entity, nil
}

// Update mengganti seluruh _data entitas.
func (s *BaseService) Update(ctx context.Context, sc scope.Scope, id uuid.UUID, data map[string]any) (*BaseDomain, error) {
	if err := s.desc.Validate(data); err != nil {
		return nil, err
	}
	rels := s.desc.DefaultRels()

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	enriched, err := s.loadAutoloadData(ctx, sc, rels, data)
	if err != nil {
		return nil, err
	}

	entity, err := s.repo.Update(ctx, tx, sc, id, enriched)
	if err != nil {
		return nil, fmt.Errorf("update %s %s: %w", s.desc.TableName(), id, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.emitSync(ctx, sc, id.String(), ActionUpdate, nil)
	return entity, nil
}

// Patch melakukan partial update pada _data.
func (s *BaseService) Patch(ctx context.Context, sc scope.Scope, id uuid.UUID, partial map[string]any) (*BaseDomain, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	entity, err := s.repo.Patch(ctx, tx, sc, id, partial)
	if err != nil {
		return nil, fmt.Errorf("patch %s %s: %w", s.desc.TableName(), id, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	s.emitSync(ctx, sc, id.String(), ActionPatch, keysOf(partial))
	return entity, nil
}

// Delete melakukan soft delete entitas.
func (s *BaseService) Delete(ctx context.Context, sc scope.Scope, id uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if err := s.repo.SoftDelete(ctx, tx, sc, id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.emitSync(ctx, sc, id.String(), ActionDelete, nil)
	return nil
}

// GetByID mengambil satu entitas berdasarkan ID.
func (s *BaseService) GetByID(ctx context.Context, sc scope.Scope, id uuid.UUID) (*BaseDomain, error) {
	return s.repo.FindByID(ctx, sc, id)
}

// List mengambil daftar entitas dengan filter dan paginasi.
func (s *BaseService) List(ctx context.Context, sc scope.Scope, p FindParams) ([]BaseDomain, int64, error) {
	return s.repo.FindAll(ctx, sc, p)
}

// loadAutoloadData meng-embed data dari related entities ke dalam _data.
// Hanya untuk relasi tipe belongs_to dengan is_autoload = true.
func (s *BaseService) loadAutoloadData(ctx context.Context, sc scope.Scope, rels map[string]RelDef, data map[string]any) (map[string]any, error) {
	result := copyMap(data)
	for relName, rel := range rels {
		if !rel.IsAutoload || rel.Type != RelBelongsTo {
			continue
		}
		fkVal, ok := data[rel.FK]
		if !ok || fkVal == nil {
			continue
		}
		embedded, err := s.fetchEmbedded(ctx, sc, rel, fmt.Sprintf("%v", fkVal))
		if err != nil {
			s.log.Warn().Err(err).Str("rel", relName).Msg("autoload fetch failed, skipping")
			continue
		}
		result[relName] = embedded
	}
	return result, nil
}

// fetchEmbedded mengambil field yang di-denormalisasi dari entitas target.
func (s *BaseService) fetchEmbedded(ctx context.Context, sc scope.Scope, rel RelDef, fkValue string) (map[string]any, error) {
	id, err := uuid.Parse(fkValue)
	if err != nil {
		return nil, fmt.Errorf("invalid FK value %q: %w", fkValue, err)
	}
	relRepo := NewBaseRepository(s.db, rel.Domain)
	entity, err := relRepo.FindByID(ctx, sc, id)
	if err != nil {
		return nil, err
	}
	return extractFields(entity.Data, rel.Fields), nil
}

// emitSync mempublish SyncEvent setelah TX commit (best-effort).
func (s *BaseService) emitSync(ctx context.Context, sc scope.Scope, id, action string, changedFields []string) {
	event := SyncEvent{
		Domain: s.desc.TableName(), EntityID: id,
		TenantID: sc.TenantID, CompanyID: sc.CompanyID,
		Action: action, ChangedFields: changedFields,
	}
	if err := s.eb.Publish(ctx, event); err != nil {
		s.log.Warn().Err(err).
			Str("domain", s.desc.TableName()).Str("id", id).
			Msg("sync event publish failed — sync worker will sweep")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func extractFields(data map[string]any, fields []string) map[string]any {
	if len(fields) == 0 {
		return data
	}
	result := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := data[f]; ok {
			result[f] = v
		}
	}
	return result
}

func copyMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	maps.Copy(dst, src)
	return dst
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// FormatItem mengubah BaseDomain menjadi map yang siap di-serialize ke JSON response.
// Merge _data dengan top-level fields (id, created_at, dll).
func FormatItem(e *BaseDomain) map[string]any {
	result := make(map[string]any, len(e.Data)+6)
	maps.Copy(result, e.Data)
	result["id"] = e.ID
	result["_sync_status"] = e.SyncStatus
	result["_sync_version"] = e.SyncVersion
	result["created_at"] = e.CreatedAt
	result["updated_at"] = e.UpdatedAt
	if e.DeletedAt != nil {
		result["deleted_at"] = e.DeletedAt
	}
	return result
}

// FormatItemJSON mengubah BaseDomain menjadi json.RawMessage untuk response.
func FormatItemJSON(e *BaseDomain) (json.RawMessage, error) {
	return json.Marshal(FormatItem(e))
}
