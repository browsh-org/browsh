# Creates a font with just the █ (0x2588) for all unicode characters
# TODO:
#   Look into using: https://github.com/adobe-fonts/adobe-blank
#   It should both reduce the size of the font and support all possible UTF8 chars

import fontforge

blocks = fontforge.font()
blocks.encoding = 'UnicodeFull'

codepoint = 0x2588
glyph = blocks.createChar(codepoint)
glyph.width = 600

pen = blocks[codepoint].glyphPen()
pen.moveTo((0, -200))
pen.lineTo((0, 800))
pen.lineTo((600, 800))
pen.lineTo((600, -200))
pen.closePath()

block = blocks[codepoint].glyphname
# There's an error if you go too high in the range :/
for i in range(0x0000, 0x6000):
  if i == codepoint: continue
  glyph = blocks.createChar(i)
  glyph.width = 600
  glyph.addReference(block)

print(blocks[codepoint].foreground)
blocks.fontname = "BlockCharMono"
blocks.fullname = 'BlockCharMono'
blocks.familyname = 'BlockCharMono'

# Fontforge's WOFF output doesn't seem to work. No matter, this isn't for an actual
# remote production website. The font is served locally from the extension and doesn't
# even need to look good.
blocks.generate("BlockCharMono.ttf")

