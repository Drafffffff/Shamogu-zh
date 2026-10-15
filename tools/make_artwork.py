#!/usr/bin/env python3
"""Deterministic Steam artwork from the official embedded tile sprites."""
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import re,base64,io,json
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'artwork'; OUT.mkdir(exist_ok=True)
source=(ROOT/'images.go').read_text()
tiles={name:Image.open(io.BytesIO(base64.b64decode(data))).convert('RGB') for name,data in re.findall(r'TileImgs\["([^"]+)"\] = \[\]byte\(`([^`]+)`\)',source)}
BG=(16,55,66); FG=(202,216,217); GOLD=(219,179,45); TEAL=(65,199,185)
fontpath=str(ROOT/'fonts/HarmonistPixelSC-Regular.ttf')
def font(n): return ImageFont.truetype(fontpath,max(12,(n//12)*12))
def sprite(name,scale,fg=FG,bg=BG):
 im=tiles[name]; mask=im.convert('L').point(lambda v:255 if v else 0)
 out=Image.new('RGB',im.size,bg);out.paste(fg,(0,0),mask)
 return out.resize((im.width*scale,im.height*scale),Image.Resampling.NEAREST)
def scene(size):
 w,h=size;im=Image.new('RGB',size,BG);d=ImageDraw.Draw(im)
 # Recognisable dungeon masonry and foliage, with a quiet title field.
 unit=max(2,w//640);tw,th=16*unit,24*unit
 for row,y in enumerate(range(-th,h,th)):
  for col,x in enumerate(range(0,w,tw)):
   if x < w*.48 and y > h*.14 and y<h*.85: continue
   name='map-wall' if (row%5==0 or col%9==0) else 'map-ground'
   if name not in tiles: name='map-notile'
   im.paste(sprite(name,unit,(33,76,85)),(x,y))
 return im
def text(im,s,xy,size,color=FG):
 ImageDraw.Draw(im).text(xy,s,font=font(size),fill=color,stroke_width=0)
def title(im,x,y,width):
 n=int(width/6.3)
 text(im,'SHAMOGU',(x,y),n,GOLD)
 text(im,'萨满山灵守护者',(x,y+n*1.3),int(width/13),FG)

def cover(size,poster=False):
 w,h=size;im=scene(size)
 if poster:
  title(im,int(w*.08),int(h*.07),int(w*.84))
  sc=max(2,int(w/80));player=sprite('map-player',sc,GOLD)
  im.paste(player,((w-player.width)//2,int(h*.42)))
  text(im,'图腾 · 潜行 · 净化腐化',(int(w*.12),int(h*.84)),int(w*.055),TEAL)
 else:
  title(im,int(w*.06),int(h*.20),int(w*.65))
  sc=max(2,int(h/35));player=sprite('map-player',sc,GOLD)
  im.paste(player,(int(w*.81),int(h*.30)))
 return im
cover((920,430)).save(OUT/'grid.png')
cover((600,900),True).save(OUT/'poster.png')
hero=scene((3840,1240));sc=28;player=sprite('map-player',sc,GOLD);hero.paste(player,(2900,260));hero.save(OUT/'hero.png')
logo=Image.new('RGBA',(1600,480),(0,0,0,0));title(logo,32,12,1500);logo.save(OUT/'logo.png')
icon=Image.new('RGB',(512,512),BG);ic=sprite('map-player',16,GOLD);icon.paste(ic,((512-ic.width)//2,(512-ic.height)//2));icon.save(OUT/'icon.png')
# A review sheet makes all slots and transparent-logo bounds inspectable.
review=Image.new('RGB',(1280,960),BG)
for name,box in [('grid',(0,0,640,300)),('poster',(900,0,1280,570)),('hero',(0,580,1280,960)),('logo',(0,320,900,550)),('icon',(680,0,880,200))]:
 a=Image.open(OUT/f'{name}.png');a.thumbnail((box[2]-box[0],box[3]-box[1]));review.paste(a,(box[0],box[1]),a if a.mode=='RGBA' else None)
review.save(OUT/'preview.jpg')
(OUT/'manifest.json').write_text(json.dumps({n:{'size':list(Image.open(OUT/f'{n}.png').size),'source':'Official v1.6.0 tiles; local composition','license':'CC-BY-4.0 (official itch.io asset license); attribution: anaseto (Yon)'} for n in ['grid','poster','hero','logo','icon']},ensure_ascii=False,indent=2))
