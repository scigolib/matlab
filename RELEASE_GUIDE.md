# Release Guide - Git-Flow Standard Process

**CRITICAL**: Read this guide BEFORE creating any release!

> **Universal guide** for releasing MATLAB File Reader/Writer using git-flow workflow

---

## 🔴 CRITICAL: Backup Before Any Operation

**ALWAYS create a backup before any serious operations!**

**Linux/macOS**:
```bash
# Directory backup
cp -r matlab matlab-backup-$(date +%Y%m%d-%H%M%S)

# Git bundle (portable, cross-platform)
git bundle create ../matlab-backup.bundle --all
```

**Windows (PowerShell)**:
```powershell
# Directory backup
Copy-Item -Recurse matlab "matlab-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"

# Git bundle (portable, cross-platform)
git bundle create ..\matlab-backup.bundle --all
```

**Git bundle** is recommended - portable, cross-platform, space-efficient!

**Dangerous operations (require backup)**:
- `git reset --hard`
- `git branch -D`
- `git tag -d`
- `git push -f`
- `git rebase`
- Any rollback/deletion operations

---

## 🎯 Git Flow Strategy

### Branch Structure

```
main        - Production-ready code ONLY (protected, green CI always)
  ↑
release/*   - Release candidates (RC)
  ↑
develop     - Active development (default branch for PRs)
  ↑
feature/*   - Feature branches
```

### Branch Rules

#### `main` Branch
- ✅ **ALWAYS** production-ready
- ✅ **ALWAYS** green CI (all tests passing)
- ✅ **ONLY** accepts merges from `release/*` branches
- ❌ **NEVER** commit directly to main
- ❌ **NEVER** push without green CI
- ❌ **NEVER** force push
- 🏷️ **Tags created ONLY after CI passes**

#### `develop` Branch
- Default branch for development
- Accepts feature branches
- May contain work-in-progress code
- Should pass tests, but can have warnings
- **Current default branch**

#### `release/*` Branches
- Format: `release/vX.Y.Z-alpha`, `release/vX.Y.Z-beta`, `release/vX.Y.Z`
- Created from `develop`
- Only bug fixes and documentation updates allowed
- No new features
- Merges to both `main` and `develop`

#### `feature/*` Branches
- Format: `feature/v5-parser`, `feature/add-tests`
- Created from `develop`
- Merged back to `develop` with `--squash` (1 clean commit per feature)

---

## 🔀 Merge Strategy (Git-Flow Standard)

### When to Use --squash vs --no-ff

**Use `--squash` (feature → develop)**:
```bash
# Feature branches: many WIP commits → 1 clean commit
git checkout develop
git merge --squash feature/my-feature
git commit -m "feat: implement my feature

- Component 1
- Component 2
- Component 3"
```

**Why squash features?**:
- Keeps develop history clean (5-10 commits per release)
- Prevents 100+ WIP commits cluttering develop
- Each feature = 1 logical commit
- Makes git log readable

**Use `--no-ff` (release → main, main → develop)**:
```bash
# Release branches: preserve complete history
git checkout main
git merge --no-ff release/vX.Y.Z-beta

# Merge back to develop
git checkout develop
git merge --no-ff main -m "Merge release vX.Y.Z back to develop"
```

**Why --no-ff for releases?**:
- Standard git-flow practice (official workflow)
- Preserves all release preparation commits
- Allows proper tag placement
- Enables clean merge back to develop
- Shows clear release boundaries in history

**NEVER Use --squash (release → main)**:
- ❌ Breaks git-flow (main ← develop merge conflicts)
- ❌ Loses release preparation history
- ❌ Makes merge back to develop difficult
- ❌ Not standard practice

---

## 🔧 Pre-Release Validation Script

### Location
`scripts/pre-release-check.sh`

### Purpose
Runs **all quality checks locally** before creating a release, matching CI requirements exactly.

### When to Use

#### 1. Before Every Commit (Recommended)
```bash
# Quick validation before committing
bash scripts/pre-release-check.sh

# If script passes (green/yellow), safe to commit:
git add .
git commit -m "..."
git push
```

#### 2. Before Creating Release Branch (Mandatory)
```bash
# MUST pass before starting release process
bash scripts/pre-release-check.sh

# Only proceed if output shows:
# ✅ "All checks passed! Ready for release."
```

#### 3. Before Merging to Main (Mandatory)
```bash
# Final validation on release branch
git checkout release/vX.Y.Z-beta
bash scripts/pre-release-check.sh

# If errors found, fix them before merging to main
```

#### 4. After Major Changes (Recommended)
- After refactoring
- After dependency updates
- After documentation updates
- After fixing bugs

### What the Script Validates

1. **Go version**: 1.25+ required
2. **Code formatting**: `gofmt -l .` must be clean
3. **Static analysis**: `go vet ./...` must pass
4. **Build**: `go build ./...` must succeed
5. **go.mod**: `go mod verify` and `go mod tidy` check
6. **Tests**: All tests passing (with race detector)
7. **Coverage**: >70% target
8. **golangci-lint**: 0 critical issues
9. **TODO/FIXME**: 0 comments required (production standard)
10. **Documentation**: All critical files present
11. **HDF5 dependency**: Verify ../hdf5 is available

### Exit Codes

- **0 (green)**: All checks passed, ready for release
- **0 (yellow)**: Warnings present, review before release
- **1 (red)**: Errors found, MUST fix before release

---

## 📝 Release Process (Step-by-Step)

### Phase 1: Preparation (on `develop` branch)

```bash
# 1. Ensure you're on develop and it's up to date
git checkout develop
git pull origin develop

# 2. Run pre-release validation (MANDATORY)
bash scripts/pre-release-check.sh

# 3. If validation passes, proceed to create release branch
```

### Phase 2: Create Release Branch

```bash
# Format: release/vX.Y.Z-alpha, release/vX.Y.Z-beta, or release/vX.Y.Z
git checkout -b release/v0.1.0-alpha develop

# Update version references in:
# - README.md (badges, version number)
# - ROADMAP.md (current version, status)
# - CHANGELOG.md (release notes)
# - go.mod (if module version changed)
# - Any documentation with version numbers

# Create ONE commit with ALL version updates
git add -A
git commit -m "chore: prepare v0.1.0-alpha release

- Update README.md version badges
- Update ROADMAP.md current status
- Add CHANGELOG.md release notes
- Update documentation versions"

# Push release branch for CI validation
git push origin release/v0.1.0-alpha
```

### Phase 3: Wait for CI (CRITICAL)

```bash
# ⏳ WAIT for CI to be GREEN on release branch
# Check: https://github.com/scigolib/matlab/actions

# If CI fails:
# 1. Fix the issues on release branch
# 2. Commit fixes: git commit -m "fix: resolve CI issue"
# 3. Wait for CI again
# 4. Repeat until GREEN

# ✅ Only proceed when CI is GREEN
```

### Phase 4: Merge to Main

```bash
# 1. Ensure CI is GREEN on release branch
# 2. Merge to main (use --no-ff to preserve history)
git checkout main
git pull origin main
git merge --no-ff release/v0.1.0-alpha -m "Release v0.1.0-alpha

Complete release notes in CHANGELOG.md"

# 3. Push to main (WITHOUT tags yet)
git push origin main

# 4. ⏳ WAIT for CI to be GREEN on main
# Check: https://github.com/scigolib/matlab/actions
```

### Phase 5: Create Tag (ONLY After CI is Green)

```bash
# ✅ ONLY after CI is GREEN on main:

# 1. Create annotated tag
git tag -a v0.1.0-alpha -m "Release v0.1.0-alpha

MATLAB File Reader v0.1.0-alpha

Features:
- Basic v5 format support
- v7.3 format support via HDF5
- Initial release

See CHANGELOG.md for details."

# 2. Push tag (PERMANENT - cannot be changed!)
git push origin main --tags

# 3. Verify tag was created
git tag -l
git show v0.1.0-alpha
```

### Phase 6: Merge Back to Develop

```bash
# 1. Merge release back to develop
git checkout develop
git pull origin develop
git merge --no-ff main -m "Merge release v0.1.0-alpha back to develop"

# 2. Push develop
git push origin develop

# 3. Delete release branch (local and remote)
git branch -d release/v0.1.0-alpha
git push origin --delete release/v0.1.0-alpha
```

### Phase 7: Create GitHub Release

1. Go to: https://github.com/scigolib/matlab/releases/new
2. Select tag: `v0.1.0-alpha`
3. Title: `MATLAB File Reader v0.1.0-alpha`
4. Description: Copy from CHANGELOG.md
5. Check "Set as a pre-release" (for alpha/beta)
6. Click "Publish release"

---

## 📋 Release Checklist

Use this checklist for every release:

### Before Creating Release Branch
- [ ] All planned features merged to `develop`
- [ ] `bash scripts/pre-release-check.sh` passes
- [ ] CHANGELOG.md is ready
- [ ] No open blocking issues

### On Release Branch
- [ ] Update README.md version
- [ ] Update ROADMAP.md status
- [ ] Complete CHANGELOG.md
- [ ] Update all documentation
- [ ] One commit with all changes
- [ ] Push and wait for CI GREEN

### Before Merging to Main
- [ ] CI is GREEN on release branch
- [ ] All tests passing
- [ ] Code review complete (if applicable)
- [ ] Documentation reviewed

### After Merging to Main
- [ ] CI is GREEN on main
- [ ] Tag created and pushed
- [ ] GitHub release created
- [ ] Merged back to develop
- [ ] Release branch deleted

### Post-Release
- [ ] Announcement (if needed)
- [ ] Archive sprint tasks in docs/dev/
- [ ] Update project board
- [ ] Plan next release

---

## 🚨 Common Mistakes to Avoid

### ❌ DON'T
1. **Push tags before CI passes** - Tags are PERMANENT!
2. **Use --squash for release → main** - Breaks git-flow
3. **Commit directly to main** - Always use release branches
4. **Skip pre-release validation** - Will fail in CI
5. **Force push to main** - Destroys history
6. **Create multiple fix commits on main** - Fix on release branch

### ✅ DO
1. **Always run pre-release script** before creating release
2. **Wait for CI to be GREEN** before pushing tags
3. **Use --no-ff for releases** (standard git-flow)
4. **Use --squash for features** (clean history)
5. **Create backups** before dangerous operations
6. **One commit per release** on release branch

---

## 🔄 Hotfix Process (Critical Bugs in Production)

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# 2. Fix the bug
git add .
git commit -m "fix: critical production bug"

# 3. Run validation
bash scripts/pre-release-check.sh

# 4. Merge to main
git checkout main
git merge --no-ff hotfix/critical-bug
git push origin main

# 5. Wait for CI, then tag
git tag -a v0.1.1-alpha -m "Hotfix v0.1.1-alpha"
git push origin main --tags

# 6. Merge to develop
git checkout develop
git merge --no-ff main
git push origin develop

# 7. Delete hotfix branch
git branch -d hotfix/critical-bug
```

---

## 📞 Support

If you encounter issues during release:

1. Check this guide carefully
2. Run `bash scripts/pre-release-check.sh` for validation
3. Check GitHub Actions for CI errors
4. Review CONTRIBUTING.md for workflow details
5. Create an issue if something is unclear

---

**Last Updated**: 2025-11-02
**Version**: 1.0 (adapted from HDF5 library)
