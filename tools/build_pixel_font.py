#!/usr/bin/env python3
"""Build the renamed OFL font from the locally preserved Fusion Pixel source."""
from pathlib import Path
from fontTools.ttLib import TTFont
import json
from fontTools.pens.ttGlyphPen import TTGlyphPen
from PIL import Image,ImageDraw,ImageFont
import re
root=Path(__file__).resolve().parents[1]
source=root/'fonts/fusion-pixel-12px-monospaced-zh_hans.ttf'
ft=TTFont(source)
# Rare Chinese characters are rasterised once at 12px and stored as square
# contours, so they remain pixel glyphs at integer 4x scale (no smooth fallback).
text=''.join(p.read_text() for p in root.glob('*.go'))
cmap=ft.getBestCmap()
missing=sorted({ord(c) for c in text if '\u3400'<=c<='\u9fff' and ord(c) not in cmap})
order=list(ft.getGlyphOrder())
fallback=ImageFont.truetype(str(root/'fonts/NotoSansSC-Regular.ttf'),12)
for cp in missing:
 im=Image.new('L',(12,12));d=ImageDraw.Draw(im);d.text((0,0),chr(cp),font=fallback,fill=255,anchor='lt')
 pen=TTGlyphPen(None)
 for y in range(12):
  for x in range(12):
   if im.getpixel((x,y))>=96:
    left=x*100;bottom=900-y*100
    pen.moveTo((left,bottom));pen.lineTo((left,bottom+100));pen.lineTo((left+100,bottom+100));pen.lineTo((left+100,bottom));pen.closePath()
 name=f'harmonistPixel{cp:05X}'
 ft['glyf'][name]=pen.glyph();ft['hmtx'][name]=(1200,0)
 if 'vmtx' in ft:ft['vmtx'][name]=(1200,0)
 order.append(name)
 for table in ft['cmap'].tables:
  if table.isUnicode():table.cmap[cp]=name
ft.setGlyphOrder(order)
values={1:'Harmonist Pixel SC',2:'Regular',3:'Harmonist Pixel SC 1.0',4:'Harmonist Pixel SC',6:'HarmonistPixelSC-Regular',16:'Harmonist Pixel SC',17:'Regular'}
for record in ft['name'].names:
 if record.nameID in values:record.string=values[record.nameID].encode(record.getEncoding(),errors='replace')
ft.save(root/'fonts/HarmonistPixelSC-Regular.ttf')
(root/'fonts/font-source.json').write_text(json.dumps({'source':'Fusion Pixel 12px monospaced zh-Hans','upstream':'https://github.com/TakWolf/fusion-pixel-font','license':'SIL OFL 1.1','pixel_size':12,'render_size':48,'pixel_scale':4,'glyph_count':len(ft.getBestCmap()),'local_source':'Preserved font from DCSS project','pixel_fallback_characters':[chr(cp) for cp in missing],'fallback_method':'Noto Sans SC rasterised at 12px into square contours'},indent=2))
