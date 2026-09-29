"""Local documentation inventory, links and diagram extraction; no model/network."""
import hashlib
import json
from pathlib import Path
import re
import sys

OUTPUT = Path(__file__).resolve().parent
ROOT = OUTPUT.parents[1]
sys.path.insert(0, str(ROOT / 'pipeline/scripts'))
from check_docs import check_document

folders = ('docs', 'docs/runbooks', 'correcao/docs', 'correcao/avaliacao',
           'correcao/adapters/python-3.13-stdlib', 'deploy/linux', 'negocio',
           'negocio/juridico', 'negocio/modelos', 'negocio/operacao',
           'pipeline/knowledge/hipporag', 'landing-page')
files = {ROOT/'README.md', ROOT/'pipeline/README.md', ROOT/'correcao/README.md'}
for folder in folders:
    path = ROOT / folder
    if path.is_symlink() or not path.is_dir():
        raise SystemExit('Invalid documentation directory: ' + folder)
    files.update(path.glob('*.md'))
errors = []
diagrams = []
inventory = []
for path in sorted(files):
    name = path.relative_to(ROOT).as_posix()
    found = check_document(ROOT, path)
    errors.extend(found)
    if found:
        continue
    text = path.read_text(encoding='utf-8')
    blocks = list(re.finditer(r'^```mermaid\s*\n(.*?)^```\s*$', text, re.M | re.S))
    if len(blocks) != len(re.findall(r'^```mermaid\s*$', text, re.M)):
        errors.append(name + ': unclosed Mermaid block')
    for index, block in enumerate(blocks, 1):
        diagrams.append({'source': name, 'number': index, 'text': block.group(1)})
    inventory.append({'path': name, 'sha256': hashlib.sha256(path.read_bytes()).hexdigest(),
                      'mermaid_blocks': len(blocks)})
report = {'status': 'failed' if errors else 'passed', 'documents': len(files),
          'diagrams': len(diagrams), 'errors': errors, 'files': inventory,
          'legal_approval': False, 'production_approval': False,
          'note': 'Current local snapshot, not a signature or alteration of historical evidence.'}
(OUTPUT/'documents.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
(OUTPUT/'diagrams.json').write_text(json.dumps(diagrams, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
print(json.dumps({k:report[k] for k in ('status','documents','diagrams','errors')}, ensure_ascii=False))
raise SystemExit(bool(errors))
