# Phase 2 Testing Checklist

## Connection Pool Tests

- [ ] Pool creates successfully with valid config
- [ ] Pool fails gracefully with invalid config
- [ ] Pool handles connection timeout
- [ ] Ping succeeds on healthy connection
- [ ] Ping fails on closed connection
- [ ] Query returns correct results
- [ ] QueryRow returns single row
- [ ] Connection string builds correctly with all fields
- [ ] Connection string handles missing password

## Connection Manager Tests

- [ ] Manager initializes with empty connections
- [ ] Connect establishes new connection
- [ ] Connect fails with invalid credentials
- [ ] Disconnect closes connection
- [ ] GetActive returns active connection
- [ ] GetActive fails when no active connection
- [ ] SetActive switches active connection
- [ ] GetAll returns all connections
- [ ] Ping updates connection state

## Auto-Discovery Tests

### Port Scanner
- [ ] Scans default ports (5432-5435)
- [ ] Detects running PostgreSQL instance
- [ ] Skips unavailable ports
- [ ] Respects timeout
- [ ] Handles context cancellation

### Environment Parser
- [ ] Reads PGHOST, PGPORT, PGDATABASE
- [ ] Returns nil when no env vars set
- [ ] Uses defaults for missing values
- [ ] Parses port correctly

### pgpass Parser
- [ ] Parses valid .pgpass file
- [ ] Skips comment lines
- [ ] Skips invalid lines
- [ ] Returns empty list when file doesn't exist
- [ ] FindPassword matches wildcards
- [ ] FindPassword returns correct password

### Discovery Coordinator
- [ ] Combines all discovery methods
- [ ] Deduplicates instances
- [ ] Prioritizes sources correctly
- [ ] Handles context cancellation

## Metadata Queries Tests

- [ ] ListDatabases returns all databases
- [ ] ListSchemas filters system schemas
- [ ] ListTables returns tables for schema
- [ ] GetTableRowCount returns estimate

## UI Component Tests

- [ ] ConnectionDialog renders in discovery mode
- [ ] ConnectionDialog renders in manual mode
- [ ] MoveSelection navigates instances
- [ ] MoveSelection navigates form fields
- [ ] GetSelectedInstance returns correct instance
- [ ] GetManualConfig builds correct config

## Integration Tests

- [ ] App opens connection dialog with 'c'
- [ ] Connection dialog shows discovered instances
- [ ] Selecting instance attempts connection
- [ ] Manual mode allows input
- [ ] ESC closes dialog
- [ ] Help shows connection keys

## Manual Testing

1. **No PostgreSQL Running**
   - [ ] Discovery shows "discovering..." message
   - [ ] Manual mode works
   - [ ] Connection fails gracefully with error message

2. **PostgreSQL on Default Port**
   - [ ] Discovery finds localhost:5432
   - [ ] Shows "Port Scan" as source
   - [ ] Connection succeeds

3. **Multiple PostgreSQL Instances**
   - [ ] Discovery finds all instances
   - [ ] Lists ports correctly
   - [ ] Can connect to each

4. **Environment Variables Set**
   - [ ] Discovery shows environment instance
   - [ ] Prioritizes environment over port scan

5. **.pgpass File Present**
   - [ ] Discovery shows .pgpass instances
   - [ ] Password auto-filled from .pgpass
   - [ ] Connection succeeds without password prompt

6. **Connection States**
   - [ ] Shows "connecting..." while connecting
   - [ ] Shows "connected" with timestamp
   - [ ] Shows error message on failure
   - [ ] Ping updates last ping time
