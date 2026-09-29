"""Render the canonical one-page Markdown. No independent copy of the offer."""
from pathlib import Path
import re
from xml.sax.saxutils import escape
from reportlab.lib import colors
from reportlab.lib.enums import TA_LEFT
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle
from pypdf import PdfReader

root = Path(__file__).resolve().parents[2]
source = root / "negocio/oferta/OFERTA_MVP_UMA_PAGINA.md"
out = root / "negocio/oferta/output/pdf/OFERTA_MVP_UMA_PAGINA.pdf"
out.parent.mkdir(parents=True, exist_ok=True)
styles = {
    "title": ParagraphStyle("title", fontName="Helvetica-Bold", fontSize=21, leading=24, textColor=colors.HexColor("#202123"), spaceAfter=11),
    "heading": ParagraphStyle("heading", fontName="Helvetica-Bold", fontSize=10.3, leading=13, textColor=colors.HexColor("#a82e3d"), spaceBefore=10, spaceAfter=4),
    "body": ParagraphStyle("body", fontName="Helvetica", fontSize=9.1, leading=12.1, alignment=TA_LEFT, spaceAfter=6, textColor=colors.HexColor("#303336")),
    "small": ParagraphStyle("small", fontName="Helvetica", fontSize=8, leading=10.5, spaceAfter=4, textColor=colors.HexColor("#52555a")),
}
def inline(text):
    text = escape(text.replace("–", "-"))
    text = re.sub(r"\*\*(.+?)\*\*", r"<b>\1</b>", text)
    def link(match):
        label, target = match.groups()
        if target.startswith("https://"):
            return f'<link href="{target}" color="#a82e3d">{label}</link>'
        return label
    return re.sub(r"\[([^]]+)\]\(([^)]+)\)", link, text)

story = []
table_rows = []
def flush_table():
    if not table_rows:
        return
    table = Table([[Paragraph(inline(cell), styles["small"]) for cell in row] for row in table_rows], colWidths=[67, 155, 301], hAlign="LEFT")
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#f4e9eb")),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("LEFTPADDING", (0, 0), (-1, -1), 7), ("RIGHTPADDING", (0, 0), (-1, -1), 7),
        ("TOPPADDING", (0, 0), (-1, -1), 6), ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
        ("LINEBELOW", (0, 0), (-1, -1), .3, colors.HexColor("#dedcda")),
    ]))
    story.extend([table, Spacer(1, 7)])
    table_rows.clear()

for line in source.read_text(encoding="utf-8").splitlines():
    if line.startswith("|"):
        if not re.match(r"\|\s*---", line):
            table_rows.append([cell.strip() for cell in line.strip("|").split("|")])
        continue
    flush_table()
    if not line.strip():
        continue
    if line.startswith("# "):
        story.append(Paragraph(inline(line[2:]), styles["title"]))
    elif line.startswith("## "):
        story.append(Paragraph(inline(line[3:]), styles["heading"]))
    else:
        style = "small" if line.startswith(("Decisão de produto", "**Base do preço")) else "body"
        story.append(Paragraph(inline(line), styles[style]))
flush_table()
def footer(canvas, _doc):
    canvas.setStrokeColor(colors.HexColor("#a82e3d"))
    canvas.line(36, 30, A4[0] - 36, 30)
    canvas.setFont("Helvetica", 7)
    canvas.setFillColor(colors.HexColor("#666666"))
    canvas.drawString(36, 19, "Vexkeep | Oferta inicial de 26/09/2026 | Proposta sujeita a escopo e homologação")
    canvas.drawRightString(A4[0] - 36, 19, "1 / 1")
SimpleDocTemplate(str(out), pagesize=A4, rightMargin=36, leftMargin=36, topMargin=34, bottomMargin=40, title="Vexkeep - Oferta inicial", author="Vexkeep").build(story, onFirstPage=footer)
reader = PdfReader(out)
assert len(reader.pages) == 1, f"Expected one page, got {len(reader.pages)}"
for required in ("1.350,00", "449,10", "5.389,20", "Quem compra", "O que recebe"):
    assert required in reader.pages[0].extract_text(), required
print(f"PDF validated: 1 page, required text present. {out}")
