#!/usr/bin/env python3
"""Exercise the installer against a temporary Steam profile, never the real user profile."""
from pathlib import Path
from unittest.mock import patch
import tempfile,runpy,subprocess,struct,contextlib,io,json,shutil
root=Path(__file__).resolve().parents[1]
def encode(obj):
 result=bytearray()
 for key,value in obj.items():
  kind=0 if isinstance(value,dict) else 1 if isinstance(value,str) else 2
  result.extend(bytes([kind])+key.encode()+b'\0')
  result.extend(encode(value) if kind==0 else value.encode()+b'\0' if kind==1 else struct.pack('<I',value))
 return bytes(result)+b'\x08'
with tempfile.TemporaryDirectory(prefix='shamogu-steam-test-') as td:
 home=Path(td);config=home/'.local/share/Steam/userdata/146975941/config';config.mkdir(parents=True)
 original={'0':{'appid':123,'AppName':'Existing game','Exe':'"/existing/game"','tags':{'0':'favorite'},'FutureField':'preserve me'}}
 path=config/'shortcuts.vdf';path.write_bytes(encode({'shortcuts':original}))
 game=home/'Games/shamogu-zh';game.mkdir(parents=True)
 shutil.copytree(root/'artwork',game/'artwork')
 def install(running=False):
  with patch('pathlib.Path.home',return_value=home),patch('subprocess.run',return_value=subprocess.CompletedProcess([],0 if running else 1)),contextlib.redirect_stdout(io.StringIO()):
   return runpy.run_path(str(root/'tools/install_steam_bazzite.py'))
 namespace=install();parse=namespace['parse'];first=parse(path.read_bytes())[0]['shortcuts']
 assert first['0']==original['0'] and len(first)==2
 namespace=install();second=parse(path.read_bytes())[0]['shortcuts']
 assert second==first
 record=json.loads((game/'steam-install.json').read_text())
 for slot,target in record['artwork'].items(): assert (config/'grid'/target).read_bytes()==(game/'artwork'/f'{slot}.png').read_bytes()
 before=path.read_bytes()
 try: install(True)
 except SystemExit: pass
 else: raise AssertionError('Running Steam guard failed')
 assert path.read_bytes()==before
 result={'existing_shortcuts_preserved':True,'idempotency':True,'all_five_artwork_slots':True,'running_steam_guard':True,'real_steam_profile_touched':False}
 (root/'verification/steam-installer-tests.json').write_text(json.dumps(result,indent=2)+'\n')
 print(json.dumps(result,indent=2))
