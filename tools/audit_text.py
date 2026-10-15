#!/usr/bin/env python3
from pathlib import Path
import subprocess,json,re,tempfile
root=Path(__file__).resolve().parents[1]
subprocess.run(['go','build','-o','/tmp/shamogu-strings',str(root/'tools/strings_dump.go')],check=True)
original=Path(tempfile.mkdtemp(prefix='shamogu-en-'))
files=[p for p in root.glob('*.go') if p.name not in ['images.go','chinese_font.go'] and not p.name.endswith('_test.go')]
for p in files:
 r=subprocess.run(['git','show','v1.6.0:'+p.name],cwd=root,capture_output=True)
 if r.returncode==0:(original/p.name).write_bytes(r.stdout)
old=json.loads(subprocess.check_output(['/tmp/shamogu-strings']+[str(original/p.name) for p in files if (original/p.name).exists()]))
new=json.loads(subprocess.check_output(['/tmp/shamogu-strings']+[str(p) for p in files if (original/p.name).exists()]))
# Position within each function remains stable for pure string translations.
def index(rows):
 out={};counts={}
 for r in rows:
  base=(r['File'],r['Function'],r['Context']);i=counts.get(base,0);counts[base]=i+1;out[base+(i,)]=r
 return out
a,b=index(old),index(new)
changed=[];unchanged=[];mismatch=[]
for key,row in a.items():
 if key not in b:continue
 nr=b[key]
 if row['Text']!=nr['Text']:
  changed.append({**row,'Translation':nr['Text']})
  placeholders=lambda s: re.findall(r'%(?:\[[0-9]+\])?[-+#0 ]*\d*(?:\.\d+)?[a-zA-Z%]',s)
  if sorted(placeholders(row['Text']))!=sorted(placeholders(nr['Text'])):mismatch.append({**row,'Translation':nr['Text']})
 elif re.search('[A-Za-z]{2}',row['Text']) and not (row['Function']=='' and re.match(r'[a-z]+\.[a-z]+/',row['Text'])):
  unchanged.append(row)
out={'baseline':'v1.6.0','changed_strings':len(changed),'unchanged_candidates':len(unchanged),'format_mismatches':mismatch,'unchanged':unchanged,'changed':changed}
(root/'verification/text-audit.json').write_text(json.dumps(out,ensure_ascii=False,indent=2))
print(json.dumps({k:v for k,v in out.items() if k not in ['unchanged','changed']},ensure_ascii=False,indent=2))
