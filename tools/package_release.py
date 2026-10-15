#!/usr/bin/env python3
from pathlib import Path
import shutil,tarfile,hashlib,json
root=Path(__file__).resolve().parents[1];out=root/'output/release';bundle=out/'Shamogu-1.6.0-zh.1-bazzite-x86_64';shutil.rmtree(bundle,ignore_errors=True);bundle.mkdir(parents=True,exist_ok=True)
(bundle/'game').mkdir(exist_ok=True)
shutil.copy2(root/'output/linux/shamogu-zh',bundle/'game/shamogu-zh')
shutil.copy2(root/'tools/launch-shamogu-zh',bundle/'game/launch-shamogu-zh')
shutil.copytree(root/'artwork',bundle/'artwork',dirs_exist_ok=True)
for name in ['LICENSE','README.zh-CN.md']:
 shutil.copy2(root/name,bundle/name)
shutil.copy2(root/'docs/shamogu.zh-CN.6',bundle/'shamogu.zh-CN.6')
shutil.copy2(root/'tools/install_steam_bazzite.py',bundle/'install_steam_bazzite.py')
(bundle/'fonts').mkdir(exist_ok=True)
for name in ['HarmonistPixelSC-Regular.ttf','OFL.txt','font-source.json']:
 shutil.copy2(root/'fonts'/name,bundle/'fonts'/name)
shutil.copytree(root/'fonts/fusion-pixel-licenses',bundle/'fonts/fusion-pixel-licenses',dirs_exist_ok=True)
files={str(p.relative_to(bundle)):hashlib.sha256(p.read_bytes()).hexdigest() for p in bundle.rglob('*') if p.is_file()}
(bundle/'SHA256.json').write_text(json.dumps(files,ensure_ascii=False,indent=2))
archive=out/(bundle.name+'.tar.gz')
with tarfile.open(archive,'w:gz') as tar:tar.add(bundle,arcname=bundle.name)
print(archive)
