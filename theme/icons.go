package theme

import "fyne.io/fyne/v2"

func svgIcon(name, paths string) *fyne.StaticResource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#e7eaf0" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`
	return fyne.NewStaticResource(name+".svg", []byte(svg))
}

var (
	IconVault = svgIcon("vault",
		`<rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="12" cy="12" r="3.2"/><path d="M12 8.8V7.5M12 16.5v-1.3M15.2 12H16.5M7.5 12h1.3"/>`)

	IconKey = svgIcon("key",
		`<circle cx="8.5" cy="14.5" r="3.5"/><path d="M11 12 21 2M17 6l2 2M14 9l2 2"/>`)

	IconWand = svgIcon("wand",
		`<path d="M15 4V2M15 16v-2M8 9h2M20 9h2M17.8 11.8l1.4 1.4M12.2 6.2l1.4 1.4M17.8 6.2l-1.4 1.4M3 21l9-9"/>`)

	IconShieldCheck = svgIcon("shield-check",
		`<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/>`)

	IconSearch = svgIcon("search",
		`<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>`)

	IconSettings = svgIcon("settings",
		`<circle cx="12" cy="12" r="2.6"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1Z"/>`)

	IconLock = svgIcon("lock",
		`<rect x="4" y="11" width="16" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>`)

	IconLockOpen = svgIcon("lock-open",
		`<rect x="4" y="11" width="16" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 7.5-2"/>`)

	IconEye = svgIcon("eye",
		`<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="2.6"/>`)

	IconEyeOff = svgIcon("eye-off",
		`<path d="M9.9 5.1A10 10 0 0 1 12 5c6.5 0 10 7 10 7a14 14 0 0 1-3 3.7M6.5 7.2A14 14 0 0 0 2 12s3.5 7 10 7c1.9 0 3.5-.5 5-1.3M3 3l18 18M10 10.5a2.6 2.6 0 0 0 3.7 3.7"/>`)

	IconCopy = svgIcon("copy",
		`<rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>`)

	IconTrash = svgIcon("trash",
		`<path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6M10 11v6M14 11v6"/>`)

	IconPlus = svgIcon("plus",
		`<path d="M12 5v14M5 12h14"/>`)

	IconEdit = svgIcon("edit",
		`<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.4 2.6a2 2 0 0 1 2.8 2.8L12 14.6 8 16l1.4-4Z"/>`)

	IconCard = svgIcon("card",
		`<rect x="2" y="5" width="20" height="14" rx="2"/><path d="M2 10h20M6 15h4"/>`)

	IconNote = svgIcon("note",
		`<path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9Z"/><path d="M14 3v6h6M8 13h8M8 17h5"/>`)

	IconRefresh = svgIcon("refresh",
		`<path d="M3 12a9 9 0 0 1 15.5-6.3L21 8M21 3v5h-5M21 12a9 9 0 0 1-15.5 6.3L3 16M3 21v-5h5"/>`)

	IconDownload = svgIcon("download",
		`<path d="M12 3v12M7 10l5 5 5-5M5 21h14"/>`)

	IconUpload = svgIcon("upload",
		`<path d="M12 17V5M7 10l5-5 5 5M5 21h14"/>`)

	IconPanelLeftClose = svgIcon("panel-left-close",
		`<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16M15 10l-2 2 2 2"/>`)

	IconPanelLeftOpen = svgIcon("panel-left",
		`<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/>`)

	IconAtom = svgIcon("atom",
		`<circle cx="12" cy="12" r="1.4"/><ellipse cx="12" cy="12" rx="9" ry="3.6"/><ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(60 12 12)"/><ellipse cx="12" cy="12" rx="9" ry="3.6" transform="rotate(120 12 12)"/>`)

	IconAlertTriangle = svgIcon("alert-triangle",
		`<path d="M10.3 3.7 2.6 17a2 2 0 0 0 1.7 3h15.4a2 2 0 0 0 1.7-3L13.7 3.7a2 2 0 0 0-3.4 0Z"/><path d="M12 9v4M12 17v.01"/>`)

	IconInfo = svgIcon("info",
		`<circle cx="12" cy="12" r="9"/><path d="M12 8v.01M11 12h1v5h1"/>`)

	IconFace = svgIcon("face",
		`<rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="9" cy="10" r="0.8" fill="#e7eaf0" stroke="none"/><circle cx="15" cy="10" r="0.8" fill="#e7eaf0" stroke="none"/><path d="M9 15.2c.9.9 2 1.3 3 1.3s2.1-.4 3-1.3"/>`)

	IconChevronLeft = svgIcon("chevron-left",
		`<path d="m15 6-6 6 6 6"/>`)

	IconChevronRight = svgIcon("chevron-right",
		`<path d="m9 6 6 6-6 6"/>`)

	IconCheck = svgIcon("check",
		`<path d="m5 12 5 5 9-11"/>`)

	IconExternalLink = svgIcon("external-link",
		`<path d="M15 3h6v6M21 3l-9 9M19 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h6"/>`)

	IconClock = svgIcon("clock",
		`<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 3"/>`)

	IconQRCode = svgIcon("qr-code",
		`<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="3" height="3"/><rect x="18" y="18" width="3" height="3"/><rect x="14" y="18" width="3" height="3"/>`)

	IconFile = svgIcon("file",
		`<path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9Z"/><path d="M14 3v6h6"/>`)

	IconFileUp = svgIcon("file-up",
		`<path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9Z"/><path d="M14 3v6h6M12 12v6M9 15l3-3 3 3"/>`)

	IconFolder = svgIcon("folder",
		`<path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9L9.2 3.6a2 2 0 0 0-1.7-.9H4a2 2 0 0 0-2 2v13.4A2 2 0 0 0 4 20Z"/>`)
)
