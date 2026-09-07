#!/usr/bin/env python3
"""Actual-binary component preset, adoption and consented migration acceptance."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

binary = os.environ['E2E_BINARY']
source = Path(os.environ['MEMORY_BANK_COMPONENT_SOURCE']).resolve()
legacy = Path(os.environ['MEMORY_BANK_LEGACY_SOURCE']).resolve()
source_ref = subprocess.check_output(['git', '-C', str(source), 'rev-parse', 'HEAD'], text=True).strip()
legacy_ref = 'f1f04de843aef45a2425d4a7351d577bbf89e940'


def run(*args, success=True):
    result = subprocess.run([binary, *map(str, args)], capture_output=True, text=True)
    if (result.returncode == 0) != success:
        raise AssertionError(f'{args}: exit {result.returncode}\n{result.stdout}\n{result.stderr}')
    return result.stdout


def input_args(root, old=False):
    return ['--repo-root', root, '--source', legacy if old else source,
            '--source-ref', legacy_ref if old else source_ref,
            '--template-version', 'legacy' if old else 'candidate']


def snapshot(root):
    return {str(p.relative_to(root)): (p.read_bytes(), p.stat().st_mode & 0o777)
            for p in root.rglob('*') if p.is_file()}


run('capabilities', '--require', 'components/v1', '--require', 'adoption/v1')
with tempfile.TemporaryDirectory(prefix='component-e2e-') as temporary:
    workspace = Path(temporary)
    for preset in ('core', 'docs', 'full', 'legacy'):
        root = workspace / preset
        root.mkdir()
        run('init', *input_args(root), '--preset', preset)
        run('doctor', '--repo-root', root)
        run('lint', '--repo-root', root)
        before = snapshot(root)
        run('pull', *input_args(root))
        assert before == snapshot(root), f'{preset}: repeat changed tree'
        if preset != 'core':
            target = 'memory-bank/features/FT-710/brief.md'
            run('document', 'create', '--repo-root', root, '--type', 'feature', '--path', target)
            assert 'flow_contract:' not in (root / target).read_text()
            if preset in ('full', 'legacy'):
                text = (root / target).read_text().replace('status: draft\n', 'status: draft\ndelivery_status: planned\n', 1)
                (root / target).write_text(text)
                run('document', 'adopt', '--repo-root', root, '--path', target,
                    '--contract', 'legacy/f1f04de/feature/v1')
                run('doctor', '--repo-root', root)
        print(f'{preset}: actual-binary init, audit, repeat and applicable document lifecycle passed')
    for adapter in ('bootstrap', 'codex', 'start-issue', 'symphony'):
        root = workspace / ('adapter-' + adapter)
        root.mkdir()
        run('init', *input_args(root), '--preset', 'core', '--adapter', adapter)
        lock = json.loads((root / 'memory-bank/.lock').read_text())
        assert adapter in lock['installation']['adapters']
        assert 'flows' in lock['installation']['components']
        run('doctor', '--repo-root', root)
    root = workspace / 'migration'
    root.mkdir()
    run('init', *input_args(root, old=True))
    before = snapshot(root)
    run('pull', *input_args(root), success=False)
    assert before == snapshot(root)
    preview = json.loads(run('pull', *input_args(root), '--migrate-components', '--dry-run', '--json'))
    assert before == snapshot(root)
    digest = preview['migration_plan_digest']
    run('pull', *input_args(root), '--migrate-components', '--migration-plan-digest', digest)
    run('doctor', '--repo-root', root)
    assert json.loads((root / 'memory-bank/.lock').read_text())['schema_version'] == 2
    print('migration: actual-binary preview consent, exact digest and schema-2 audit passed')
