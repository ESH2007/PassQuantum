// All UI copy, keyed by language. Strings with {placeholders} are assembled
// by <RichText>; the placeholder names map to styled nodes in each component.

import { REPO_URL } from './config'

const t = {
  es: {
    nav: {
      links: [
        { href: '#features',     label: 'Características' },
        { href: '#architecture', label: 'Arquitectura'   },
        { href: '#transparency', label: 'Transparencia'  },
      ],
    },

    hero: {
      badge: 'Criptografía Post-Cuántica · NIST FIPS 203/204',
      // {accent} is highlighted, {underline} gets the underline flourish
      headline: 'Diseñado para {accent}\nProtegido contra {underline} hoy.',
      headlineAccent: 'el mañana.',
      headlineUnderline: 'computadores cuánticos',
      // {aes} {mlkem} {mldsa} are injected as monospace tech labels
      sub: 'Triple capa de encriptación: {aes} para payloads, {mlkem} para encapsulación de clave, y {mldsa} para firma digital. Biometría continua pasiva con MediaPipe.',
      cta1: 'Ver código en GitHub',
      cta2: 'Explorar características',
      download: {
        label:       'Descargar para',
        options:     { linux: 'Linux (amd64)', windows: 'Windows (amd64)', 'darwin-dmg': 'macOS (.dmg)', 'darwin-app': 'macOS (.app)' },
        unavailable: 'Próximamente para macOS',
      },
    },

    features: {
      tag:   '// características',
      title: 'Seguridad sin compromisos',
      sub:   'Cada capa fue elegida con criterio. Sin dependencias externas de criptografía, sin backends en la nube, sin concesiones.',
      items: [
        {
          title: 'Encriptación Híbrida',
          sub:   'Go Crypto · Nativo',
          desc:  'Triple capa criptográfica sobre librerías nativas de Go. Resistente a ataques de computadores cuánticos con los estándares NIST FIPS 203 y 204.',
        },
        {
          title: 'Biometría Continua Pasiva',
          sub:   'MediaPipe Face Landmarker',
          desc:  'Monitoreo facial en tiempo real mediante un proceso Python independiente. Auto-bloqueo tras 5 s sin cara reconocida. Liveness detection por detección de parpadeo.',
        },
        {
          title: 'Zero-Knowledge Local-First',
          sub:   '.pqdb · POSIX',
          desc:  'Sin telemetría, sin nube, sin red. Todos los datos permanecen en tu máquina en archivos .pqdb cifrados. Tu privacidad no es negociable.',
        },
      ],
    },

    architecture: {
      tag:     '// arquitectura',
      title:   'Stack tecnológico',
      sub:     'Cada componente fue seleccionado por su madurez, auditabilidad y resistencia cuántica. Criptografía exclusivamente de librerías nativas de Go.',
      headers: ['Capa', 'Tecnología', 'Módulo', 'Detalle', 'Estado'],
      status:  'ACTIVO',
      layerLabel: 'CAPA',
      stackItems: [
        { name: 'Runtime',             role: 'Desktop UI & orquestación'          },
        { name: 'Cifrado de vault',    role: 'Cifrado de payload del vault'        },
        { name: 'Post-Quantum KEM',    role: 'Encapsulación de clave por ítem'     },
        { name: 'Post-Quantum DSA',    role: 'Verificador de perfil de app'        },
        { name: 'Derivación de clave', role: 'Derivación de claves de contraseña'  },
        { name: 'Almacenamiento',      role: 'Archivos de vault cifrados'          },
      ],
      flowTag:  '// flujo de cifrado por ítem',
      flowNote: 'Cada ítem tiene su propia encapsulación ML-KEM independiente del cifrado externo del vault. Doble capa: vault-level AES-GCM + item-level AES-GCM con shared secret Kyber.',
    },

    footer: {
      transparencyTag: '// transparencia del proyecto',
      // {indie} {nativeCrypto} {notAudited} are emphasized; {lib1} {lib2} are code
      transparency: 'PassQuantum es un {indie}, desarrollado con asistencia de IA para el scaffolding y la UI, pero con {nativeCrypto} ({lib1} & {lib2}). El diseño criptográfico fue revisado manualmente. {notAudited} Úsalo bajo tu propio criterio y respalda siempre tus claves.',
      transparencyIndie: 'proyecto indie independiente',
      transparencyNativeCrypto: 'toda la criptografía manejada exclusivamente por librerías nativas de Go',
      transparencyNotAudited: 'No ha sido auditado por terceros.',
      cards: [
        { label: '✓ Crypto nativa', detail: 'Solo librerías de Go stdlib. Sin wrappers de terceros ni dependencias NPM de cripto.' },
        { label: '✓ Local-first',   detail: 'Sin servidores, sin telemetría, sin dependencias de red. Tus datos no salen de tu máquina.' },
        { label: '✓ Open Source',   detail: 'Código 100% público en GitHub. Léelo, auditalo y contribuye si quieres.' },
      ],
      navLinks: [
        { href: '#features',     label: 'Características' },
        { href: '#architecture', label: 'Arquitectura' },
        { href: REPO_URL, label: 'GitHub ↗', external: true },
      ],
      madeWith:   'Hecho con',
      by:         'por',
      disclaimer: 'PassQuantum no ofrece garantías. Haz siempre una copia de seguridad de tus claves.',
    },
  },

  en: {
    nav: {
      links: [
        { href: '#features',     label: 'Features'     },
        { href: '#architecture', label: 'Architecture' },
        { href: '#transparency', label: 'Transparency' },
      ],
    },

    hero: {
      badge: 'Post-Quantum Cryptography · NIST FIPS 203/204',
      headline: 'Built for {accent}\nProtected against {underline} today.',
      headlineAccent: 'tomorrow.',
      headlineUnderline: 'quantum computers',
      sub: 'Triple encryption layer: {aes} for payloads, {mlkem} for key encapsulation, and {mldsa} for digital signatures. Continuous passive biometrics with MediaPipe.',
      cta1: 'View code on GitHub',
      cta2: 'Explore features',
      download: {
        label:       'Download for',
        options:     { linux: 'Linux (amd64)', windows: 'Windows (amd64)', 'darwin-dmg': 'macOS (.dmg)', 'darwin-app': 'macOS (.app)' },
        unavailable: 'macOS — coming soon',
      },
    },

    features: {
      tag:   '// features',
      title: 'Security without compromise',
      sub:   'Every layer was chosen with care. No external crypto dependencies, no cloud backends, no trade-offs.',
      items: [
        {
          title: 'Hybrid Encryption',
          sub:   'Go Crypto · Native',
          desc:  'Triple cryptographic layer built on native Go libraries. Resistant to quantum computer attacks using the NIST FIPS 203 and 204 standards.',
        },
        {
          title: 'Continuous Passive Biometrics',
          sub:   'MediaPipe Face Landmarker',
          desc:  'Real-time facial monitoring via an independent Python subprocess. Auto-lock after 5 s with no recognised face. Liveness detection via blink detection.',
        },
        {
          title: 'Zero-Knowledge Local-First',
          sub:   '.pqdb · POSIX',
          desc:  'No telemetry, no cloud, no network. All data stays on your machine in encrypted .pqdb files. Your privacy is non-negotiable.',
        },
      ],
    },

    architecture: {
      tag:     '// architecture',
      title:   'Technology Stack',
      sub:     'Every component was selected for maturity, auditability, and quantum resistance. Cryptography exclusively from native Go libraries.',
      headers: ['Layer', 'Technology', 'Module', 'Detail', 'Status'],
      status:  'ACTIVE',
      layerLabel: 'LAYER',
      stackItems: [
        { name: 'Runtime',          role: 'Desktop UI & orchestration' },
        { name: 'Vault encryption', role: 'Vault payload encryption'    },
        { name: 'Post-Quantum KEM', role: 'Per-item key encapsulation'  },
        { name: 'Post-Quantum DSA', role: 'App verifier signature'      },
        { name: 'Key Derivation',   role: 'Password key derivation'     },
        { name: 'Storage',          role: 'Encrypted vault files'       },
      ],
      flowTag:  '// per-item encryption flow',
      flowNote: 'Each item has its own ML-KEM encapsulation, independent of the outer vault encryption. Double layer: vault-level AES-GCM + item-level AES-GCM with a Kyber shared secret.',
    },

    footer: {
      transparencyTag: '// project transparency',
      transparency: 'PassQuantum is an {indie}, developed with AI assistance for scaffolding and UI, but with {nativeCrypto} ({lib1} & {lib2}). The cryptographic design was manually reviewed. {notAudited} Use it at your own discretion and always back up your keys.',
      transparencyIndie: 'independent indie project',
      transparencyNativeCrypto: 'all cryptography handled exclusively by native Go libraries',
      transparencyNotAudited: 'It has not been audited by a third party.',
      cards: [
        { label: '✓ Native crypto', detail: 'Go stdlib libraries only. No third-party wrappers or NPM crypto dependencies.' },
        { label: '✓ Local-first',   detail: 'No servers, no telemetry, no network dependencies. Your data never leaves your machine.' },
        { label: '✓ Open Source',   detail: '100% public code on GitHub. Read it, audit it, contribute if you want.' },
      ],
      navLinks: [
        { href: '#features',     label: 'Features' },
        { href: '#architecture', label: 'Architecture' },
        { href: REPO_URL, label: 'GitHub ↗', external: true },
      ],
      madeWith:   'Made with',
      by:         'by',
      disclaimer: 'PassQuantum is provided as-is with no warranty. Always back up your keys.',
    },
  },
}

export default t
