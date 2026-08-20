import struct
import io
from PIL import Image, ImageDraw

def draw_discord_proxy_icon(size=256):
    img = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    scale = size / 256.0
    
    # 1. Base Squircle (Discord Blurple background)
    pad = int(8 * scale)
    radius = int(52 * scale)
    rect = [pad, pad, size - pad, size - pad]
    blurple = (88, 101, 242, 255) # #5865F2
    draw.rounded_rectangle(rect, radius=radius, fill=blurple)

    # 2. Discord Clydie Face (White)
    clydie_color = (255, 255, 255, 255)
    cx, cy = size / 2, size / 2 - 8 * scale
    
    w = 142 * scale
    h = 96 * scale
    
    bx0, by0 = cx - w/2, cy - h/2
    bx1, by1 = cx + w/2, cy + h/2
    
    # Body
    draw.ellipse([bx0, by0 + 5*scale, bx0 + 66*scale, by1], fill=clydie_color)
    draw.ellipse([bx1 - 66*scale, by0 + 5*scale, bx1, by1], fill=clydie_color)
    draw.rectangle([bx0 + 26*scale, by0 + 10*scale, bx1 - 26*scale, by1 - 10*scale], fill=clydie_color)
    draw.ellipse([cx - 46*scale, by0, cx + 46*scale, by0 + 40*scale], fill=clydie_color)
    draw.rectangle([cx - 36*scale, by0 + 10*scale, cx + 36*scale, by1 - 5*scale], fill=clydie_color)

    # Eyes
    eye_r = 14 * scale
    draw.ellipse([cx - 32*scale - eye_r, cy - 2*scale - eye_r, cx - 32*scale + eye_r, cy - 2*scale + eye_r], fill=blurple)
    draw.ellipse([cx + 32*scale - eye_r, cy - 2*scale - eye_r, cx + 32*scale + eye_r, cy - 2*scale + eye_r], fill=blurple)
    
    # Smile
    draw.arc([cx - 22*scale, cy + 10*scale, cx + 22*scale, cy + 32*scale], start=30, end=150, fill=blurple, width=max(1, int(6*scale)))

    # 3. Proxy Badge Overlay in Bottom-Right Corner
    badge_r = 46 * scale
    bcx = size - pad - badge_r + 2*scale
    bcy = size - pad - badge_r + 2*scale
    
    border_w = int(7 * scale)
    # Dark border
    draw.ellipse([bcx - badge_r - border_w, bcy - badge_r - border_w, bcx + badge_r + border_w, bcy + badge_r + border_w], fill=(49, 51, 56, 255))
    # Badge inner background
    draw.ellipse([bcx - badge_r, bcy - badge_r, bcx + badge_r, bcy + badge_r], fill=(30, 31, 34, 255))
    
    # Shield in Neon Green (#57F287)
    neon_green = (87, 242, 135, 255)
    sw = 23 * scale
    sh = 27 * scale
    
    pts = [
        (bcx - sw, bcy - sh + 4*scale),
        (bcx + sw, bcy - sh + 4*scale),
        (bcx + sw, bcy + 2*scale),
        (bcx, bcy + sh),
        (bcx - sw, bcy + 2*scale)
    ]
    draw.polygon(pts, fill=neon_green)
    
    # Inner cutout for proxy keyhole/network node
    inner_dark = (30, 31, 34, 255)
    draw.ellipse([bcx - 8*scale, bcy - 13*scale, bcx + 8*scale, bcy + 2*scale], fill=inner_dark)
    draw.polygon([(bcx - 6*scale, bcy - 4*scale), (bcx + 6*scale, bcy - 4*scale), (bcx + 8*scale, bcy + 11*scale), (bcx - 8*scale, bcy + 11*scale)], fill=inner_dark)

    return img

def create_raw_dib(image):
    """Converts a PIL RGBA Image to Windows 32-bit DIB format (BITMAPINFOHEADER + BGRA bottom-up + AND mask)"""
    w, h = image.size
    rgba_image = image.convert("RGBA")
    
    # 40-byte BITMAPINFOHEADER
    # biHeight is h * 2 (includes XOR image + AND mask)
    bih = struct.pack(
        "<LLLHHLLLLLL",
        40,          # biSize
        w,           # biWidth
        h * 2,       # biHeight
        1,           # biPlanes
        32,          # biBitCount
        0,           # biCompression (BI_RGB)
        w * h * 4,   # biSizeImage
        0, 0, 0, 0   # resolutions & colors
    )
    
    # Pixel data: bottom-to-top, BGRA
    pixels = bytearray()
    for y in range(h - 1, -1, -1):
        for x in range(w):
            r, g, b, a = rgba_image.getpixel((x, y))
            pixels.extend([b, g, r, a])
            
    # 1-bit AND mask: (row padded to 32-bit boundary)
    row_bytes = (w + 31) // 32 * 4
    and_mask = bytearray(row_bytes * h) # all 0 (fully opaque controlled by alpha)
    
    return bih + bytes(pixels) + bytes(and_mask)

def create_windows_ico(sizes=[256, 128, 64, 48, 32, 16]):
    images = [draw_discord_proxy_icon(s) for s in sizes]
    
    # 256x256 PNG for main display
    images[0].save("/home/jean/discord_proxy_launcher/icon.png", format="PNG")
    
    encoded_images = []
    for s, img in zip(sizes, images):
        if s == 256:
            # 256x256 is stored as PNG
            png_buf = io.BytesIO()
            img.save(png_buf, format="PNG")
            encoded_images.append((s, png_buf.getvalue()))
        else:
            # <= 128x128 stored as uncompressed 32-bit DIB
            dib_data = create_raw_dib(img)
            encoded_images.append((s, dib_data))
            
    # Build ICO file
    # ICONDIR: 6 bytes
    # ICONDIRENTRY: 16 bytes each
    num_icons = len(encoded_images)
    header = struct.pack("<HHH", 0, 1, num_icons)
    
    offset = 6 + num_icons * 16
    dir_entries = []
    data_blobs = []
    
    for s, data in encoded_images:
        w_byte = 0 if s == 256 else s
        h_byte = 0 if s == 256 else s
        size_data = len(data)
        
        entry = struct.pack(
            "<BBBBHHLL",
            w_byte,      # bWidth
            h_byte,      # bHeight
            0,           # bColorCount
            0,           # bReserved
            1,           # wPlanes
            32,          # wBitCount
            size_data,   # dwBytesInRes
            offset       # dwImageOffset
        )
        dir_entries.append(entry)
        data_blobs.append(data)
        offset += size_data
        
    ico_bytes = header + b"".join(dir_entries) + b"".join(data_blobs)
    
    with open("/home/jean/discord_proxy_launcher/icon.ico", "wb") as f:
        f.write(ico_bytes)
        
    print(f"Windows Standard ICO generated successfully! Total icons: {num_icons}, Size: {len(ico_bytes)} bytes")

if __name__ == "__main__":
    create_windows_ico()
