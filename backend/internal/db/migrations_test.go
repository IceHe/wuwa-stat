package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestEnsureTacetRecordSchemaBackfillsClaimCountFromRewardMode(t *testing.T) {
	database := openScriptedDB(t,
		scriptedResponse{
			kind:    "query",
			match:   "FROM information_schema.tables",
			columns: []string{"exists"},
			rows:    [][]driver.Value{{true}},
		},
		scriptedResponse{
			kind:    "query",
			match:   "table_name = 'tacet_records' AND column_name = 'claim_count'",
			columns: []string{"exists"},
			rows:    [][]driver.Value{{false}},
		},
		scriptedResponse{
			kind:  "exec",
			match: "ALTER TABLE tacet_records ADD COLUMN claim_count INTEGER NOT NULL DEFAULT 1",
		},
		scriptedResponse{
			kind:    "query",
			match:   "table_name = 'tacet_records' AND column_name = 'reward_mode'",
			columns: []string{"exists"},
			rows:    [][]driver.Value{{true}},
		},
		scriptedResponse{
			kind:  "exec",
			match: "UPDATE tacet_records SET claim_count = CASE WHEN reward_mode = 'double' THEN 2 ELSE 1 END",
		},
	)

	if err := ensureTacetRecordSchema(context.Background(), database); err != nil {
		t.Fatalf("ensureTacetRecordSchema() error = %v", err)
	}
}

func TestEnsureTacetRecordSchemaNoopWhenClaimCountExists(t *testing.T) {
	database := openScriptedDB(t,
		scriptedResponse{
			kind:    "query",
			match:   "FROM information_schema.tables",
			columns: []string{"exists"},
			rows:    [][]driver.Value{{true}},
		},
		scriptedResponse{
			kind:    "query",
			match:   "table_name = 'tacet_records' AND column_name = 'claim_count'",
			columns: []string{"exists"},
			rows:    [][]driver.Value{{true}},
		},
	)

	if err := ensureTacetRecordSchema(context.Background(), database); err != nil {
		t.Fatalf("ensureTacetRecordSchema() error = %v", err)
	}
}

func TestEnsureCreatedByUserIDColumnAddsMissingColumn(t *testing.T) {
	database := openScriptedDB(t,
		scriptedResponse{
			kind:    "query",
			match:   "column_name = 'created_by_user_id'",
			args:    []driver.Value{"ascension_records"},
			columns: []string{"exists"},
			rows:    [][]driver.Value{{false}},
		},
		scriptedResponse{
			kind:  "exec",
			match: "ALTER TABLE ascension_records ADD COLUMN created_by_user_id BIGINT",
		},
	)

	if err := ensureCreatedByUserIDColumn(context.Background(), database, "ascension_records"); err != nil {
		t.Fatalf("ensureCreatedByUserIDColumn() error = %v", err)
	}
}

func TestEnsureClaimCountColumnSkipsExistingColumn(t *testing.T) {
	database := openScriptedDB(t,
		scriptedResponse{
			kind:    "query",
			match:   "column_name = 'claim_count'",
			args:    []driver.Value{"resonance_records"},
			columns: []string{"exists"},
			rows:    [][]driver.Value{{true}},
		},
	)

	if err := ensureClaimCountColumn(context.Background(), database, "resonance_records"); err != nil {
		t.Fatalf("ensureClaimCountColumn() error = %v", err)
	}
}

type scriptedResponse struct {
	kind    string
	match   string
	args    []driver.Value
	columns []string
	rows    [][]driver.Value
}

type scriptedState struct {
	t         *testing.T
	mu        sync.Mutex
	responses []scriptedResponse
	index     int
}

type scriptedDriver struct {
	state *scriptedState
}

type scriptedConn struct {
	state *scriptedState
}

type scriptedRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

var scriptedDriverID uint64

func openScriptedDB(t *testing.T, responses ...scriptedResponse) *sql.DB {
	t.Helper()

	state := &scriptedState{t: t, responses: responses}
	name := fmt.Sprintf("scripted-db-%d", atomic.AddUint64(&scriptedDriverID, 1))
	sql.Register(name, scriptedDriver{state: state})

	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("database.Close() error = %v", err)
		}
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.index != len(state.responses) {
			t.Fatalf("consumed %d/%d scripted responses", state.index, len(state.responses))
		}
	})

	return database
}

func (d scriptedDriver) Open(name string) (driver.Conn, error) {
	return scriptedConn{state: d.state}, nil
}

func (c scriptedConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("Prepare not supported in scriptedConn")
}

func (c scriptedConn) Close() error {
	return nil
}

func (c scriptedConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("Begin not supported in scriptedConn")
}

func (c scriptedConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	response := c.state.next("query", query, args)
	return &scriptedRows{
		columns: response.columns,
		rows:    response.rows,
		index:   -1,
	}, nil
}

func (c scriptedConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.next("exec", query, args)
	return driver.RowsAffected(1), nil
}

func (c scriptedConn) Ping(_ context.Context) error {
	return nil
}

func (s *scriptedState) next(kind, query string, args []driver.NamedValue) scriptedResponse {
	s.t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.index >= len(s.responses) {
		s.t.Fatalf("unexpected %s: %s", kind, query)
	}

	response := s.responses[s.index]
	s.index++

	if response.kind != kind {
		s.t.Fatalf("response %d kind = %s, want %s", s.index-1, response.kind, kind)
	}

	if !strings.Contains(normalizeSQL(query), normalizeSQL(response.match)) {
		s.t.Fatalf("response %d query = %q, want match %q", s.index-1, normalizeSQL(query), normalizeSQL(response.match))
	}

	if response.args != nil {
		gotArgs := make([]driver.Value, len(args))
		for i, arg := range args {
			gotArgs[i] = arg.Value
		}
		if len(gotArgs) != len(response.args) {
			s.t.Fatalf("response %d args len = %d, want %d", s.index-1, len(gotArgs), len(response.args))
		}
		for i := range gotArgs {
			if gotArgs[i] != response.args[i] {
				s.t.Fatalf("response %d arg %d = %v, want %v", s.index-1, i, gotArgs[i], response.args[i])
			}
		}
	}

	return response
}

func (r *scriptedRows) Columns() []string {
	return r.columns
}

func (r *scriptedRows) Close() error {
	return nil
}

func (r *scriptedRows) Next(dest []driver.Value) error {
	r.index++
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	return nil
}

func normalizeSQL(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
