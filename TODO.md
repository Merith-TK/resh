# Development TODO

## Phase 1: Core Infrastructure ✓ (Partially Complete)
- [x] Project structure
- [x] Basic ResoLink client skeleton
- [x] WebSocket connection management
- [x] Message protocol structure
- [ ] Complete message handling
- [ ] Error handling and reconnection logic
- [ ] Unit tests for client

## Phase 2: Object Model (In Progress)
- [x] Slot type definition
- [x] Component type definition
- [ ] Member type system implementation
- [ ] Property serialization
- [ ] Property deserialization
- [ ] Type definitions for common components
- [ ] Unit tests for object model

## Phase 3: Virtual Filesystem
- [x] VFS structure
- [x] Basic path resolution
- [ ] Complete node loading logic
- [ ] Response parsing
- [ ] Cache management
- [ ] Cache invalidation strategy
- [ ] Lazy loading implementation
- [ ] Unit tests for VFS

## Phase 4: Core Commands
- [x] REPL basic structure
- [x] pwd, cd, ls commands (skeleton)
- [ ] cat command (show slot/component properties)
- [ ] tree command (hierarchical view)
- [ ] stat command (detailed metadata)
- [ ] mkdir command (create slot)
- [ ] touch command (add component)
- [ ] rm command (remove slot/component)
- [ ] Command tests

## Phase 5: REPL Features
- [x] Basic readline integration
- [x] Command history
- [x] Dynamic prompt
- [ ] Tab completion for commands
- [ ] Tab completion for paths
- [ ] Tab completion for component types
- [ ] Multi-line editing
- [ ] Syntax highlighting
- [ ] History persistence

## Phase 6: Advanced Commands
- [ ] edit command (interactive editing)
- [ ] set command (property updates)
- [ ] mv command (move/rename)
- [ ] cp command (copy hierarchies)
- [ ] find command (search)
- [ ] grep command (search in properties)
- [ ] inspect command (detailed inspector)

## Phase 7: Scripting & Automation
- [ ] Script execution mode
- [ ] Batch command files
- [ ] Variables and substitution
- [ ] Control flow (if/loop)
- [ ] Script library

## Testing & Quality
- [ ] Unit tests for all packages
- [ ] Integration tests
- [ ] Mock WebSocket server for testing
- [ ] CI/CD setup
- [ ] Code coverage reports
- [ ] Performance benchmarks

## Documentation
- [x] README
- [x] ARCHITECTURE
- [ ] API documentation (godoc)
- [ ] User guide
- [ ] Command reference
- [ ] Examples and tutorials
- [ ] Contributing guide

## Polish
- [ ] Better error messages
- [ ] Progress indicators for slow operations
- [ ] Configuration validation
- [ ] Logging system
- [ ] Debug mode
- [ ] Version command
- [ ] Update checker

## Future Features
- [ ] Component templates
- [ ] ProtoFlux visualization
- [ ] Diff and merge operations
- [ ] Remote instance support
- [ ] Plugin system
- [ ] Export/import functionality
