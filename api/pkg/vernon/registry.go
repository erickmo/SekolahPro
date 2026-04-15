package vernon

import (
	"fmt"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// ConsumerRef merepresentasikan domain yang mengonsumsi perubahan dari domain lain.
type ConsumerRef struct {
	ConsumerDomain string
	RelName        string
	RelDef         RelDef
}

// RegisteredDomain menyimpan komponen-komponen satu domain Vernon.
type RegisteredDomain struct {
	Descriptor DomainDescriptor
	Handler    *BaseHandler
	Service    *BaseService
}

// Registry menyimpan semua domain Vernon yang terdaftar.
// Menyediakan reverse dependency map untuk SyncEngine dan route mounting.
type Registry struct {
	mu         sync.RWMutex
	domains    map[string]*RegisteredDomain
	reverseMap map[string][]ConsumerRef // changedDomain → []ConsumerRef
	log        zerolog.Logger
}

// NewRegistry membuat Registry baru.
func NewRegistry(log zerolog.Logger) *Registry {
	return &Registry{
		domains:    make(map[string]*RegisteredDomain),
		reverseMap: make(map[string][]ConsumerRef),
		log:        log,
	}
}

// Register mendaftarkan domain Vernon ke registry.
// Membangun reverse dependency map dan mendeteksi circular autoload.
// Panic jika circular dependency ditemukan — ini adalah programming error.
func (r *Registry) Register(name string, d *RegisteredDomain) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.domains[name] = d
	r.buildReverseMap(name, d)

	if err := r.detectCycles(); err != nil {
		panic(fmt.Sprintf("vernon Registry: %v", err))
	}
	r.log.Info().Str("domain", name).Msg("vernon domain registered")
}

// GetDomain mengambil RegisteredDomain berdasarkan nama tabel.
func (r *Registry) GetDomain(name string) (*RegisteredDomain, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.domains[name]
	return d, ok
}

// GetConsumers mengembalikan semua domain yang mengonsumsi perubahan dari changedDomain.
func (r *Registry) GetConsumers(changedDomain string) []ConsumerRef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reverseMap[changedDomain]
}

// MountRoutes memasang route semua domain Vernon ke router.
// Setiap domain di-mount di /{tableName}/.
func (r *Registry) MountRoutes(router chi.Router) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for name, d := range r.domains {
		localName := name
		localHandler := d.Handler
		router.Route("/"+localName, func(r chi.Router) {
			localHandler.RegisterRoutes(r)
		})
		r.log.Info().Str("prefix", "/"+localName).Msg("vernon routes mounted")
	}
}

// buildReverseMap memperbarui reverseMap ketika domain baru didaftarkan.
// Untuk setiap rel dengan is_autoload = true, tambahkan consumer reference
// ke domain target sehingga SyncEngine bisa menemukan konsumer dengan cepat.
func (r *Registry) buildReverseMap(consumerName string, d *RegisteredDomain) {
	for relName, rel := range d.Descriptor.DefaultRels() {
		if !rel.IsAutoload {
			continue
		}
		ref := ConsumerRef{
			ConsumerDomain: consumerName,
			RelName:        relName,
			RelDef:         rel,
		}
		r.reverseMap[rel.Domain] = append(r.reverseMap[rel.Domain], ref)
	}
}

// GetSourceDomains mengembalikan nama semua domain yang MENJADI sumber perubahan
// (yaitu domain yang ada di reverseMap sebagai key).
// Dipakai oleh SyncEngine untuk menentukan topic apa yang perlu di-subscribe.
func (r *Registry) GetSourceDomains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	domains := make([]string, 0, len(r.reverseMap))
	for d := range r.reverseMap {
		domains = append(domains, d)
	}
	return domains
}

// detectCycles memeriksa circular autoload menggunakan DFS (white/gray/black coloring).
// Gray = sedang diproses. Jika node gray ditemukan lagi, ada cycle.
func (r *Registry) detectCycles() error {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)

	var dfs func(node string) error
	dfs = func(node string) error {
		color[node] = gray
		for _, ref := range r.reverseMap[node] {
			next := ref.ConsumerDomain
			if color[next] == gray {
				return fmt.Errorf(
					"%w: %s → %s — set is_autoload:false on one side",
					ErrCycleDetected, node, next,
				)
			}
			if color[next] == white {
				if err := dfs(next); err != nil {
					return err
				}
			}
		}
		color[node] = black
		return nil
	}

	for name := range r.domains {
		if color[name] == white {
			if err := dfs(name); err != nil {
				return err
			}
		}
	}
	return nil
}
