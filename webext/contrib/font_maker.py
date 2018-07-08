# TODO:
#   Look into using: https://github.com/adobe-fonts/adobe-blank
#   It should both reduce the size of the font and support all possible UTF8 chars

import fontforge

def generate(name, block):
    print("Generating " + name)
    # TODO:
    #   This needs to reach 0x9FCF to complete the CJK Ideographs
    #   But above around 0x7f00, we get this error:
    #   `Internal Error: Attempt to output 81854 into a 16-bit field. It will be
    #    truncated and the file may not be useful.`
    for i in range(0x0000, 0x7F00):
      if i == codepoint: continue
      glyph = blocks.createChar(i)
      glyph.width = 600
      glyph.addReference(block)

    print(blocks[codepoint].foreground)
    blocks.fontname = name
    blocks.fullname = name
    blocks.familyname = name

    # Fontforge's WOFF output doesn't seem to work. No matter, this isn't for an actual
    # remote production website. The font is served locally from the extension and doesn't
    # even need to look good.
    blocks.generate(name + '.ttf')

# A font with just the █ (0x2588) for all unicode characters
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

generate('BlockCharMono', blocks[codepoint].glyphname)

# A font with just the space character, used to hide all text
blocks = fontforge.font()
blocks.encoding = 'UnicodeFull'

codepoint = 0x2003
glyph = blocks.createChar(codepoint)
glyph.width = 600

pen = blocks[codepoint].glyphPen()
pen.moveTo((0, 0))
pen.lineTo((0, 0))
pen.closePath()

generate('BlankMono', blocks[codepoint].glyphname)
