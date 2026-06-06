; vibemd Windows NSIS Installer
; Produces a single vibemd-setup.exe that installs the app and registers .md/.markdown

!include "MUI2.nsh"

Name "vibemd"
OutFile "..\..\dist\vibemd-setup.exe"
InstallDir "$PROGRAMFILES64\vibemd"
InstallDirRegKey HKCU "Software\vibemd" ""
RequestExecutionLevel admin

; MUI pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Install"
  SetOutPath "$INSTDIR"
  File "..\..\build\bin\vibemd.exe"

  ; WebView2 bootstrapper — downloads and installs WebView2 if not present
  ; Windows 11 ships with it; most Win10 machines have it via Edge
  File "MicrosoftEdgeWebview2Setup.exe"
  ExecWait '"$INSTDIR\MicrosoftEdgeWebview2Setup.exe" /silent /install'

  ; File associations: .md and .markdown
  WriteRegStr HKCR ".md"       "" "vibemd.Document"
  WriteRegStr HKCR ".markdown" "" "vibemd.Document"
  WriteRegStr HKCR ".mdown"    "" "vibemd.Document"
  WriteRegStr HKCR "vibemd.Document" "" "Markdown Document"
  WriteRegStr HKCR "vibemd.Document\DefaultIcon" "" "$INSTDIR\vibemd.exe,0"
  WriteRegStr HKCR "vibemd.Document\shell\open\command" "" '"$INSTDIR\vibemd.exe" "%1"'

  ; Start Menu shortcut
  CreateDirectory "$SMPROGRAMS\vibemd"
  CreateShortcut "$SMPROGRAMS\vibemd\vibemd.lnk" "$INSTDIR\vibemd.exe"

  ; Uninstaller
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\vibemd" \
    "DisplayName" "vibemd"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\vibemd" \
    "UninstallString" "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\vibemd.exe"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir  "$INSTDIR"

  DeleteRegKey HKCR ".md"
  DeleteRegKey HKCR ".markdown"
  DeleteRegKey HKCR ".mdown"
  DeleteRegKey HKCR "vibemd.Document"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\vibemd"

  Delete "$SMPROGRAMS\vibemd\vibemd.lnk"
  RMDir  "$SMPROGRAMS\vibemd"
SectionEnd
