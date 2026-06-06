/**
 * FaceMesh — illustrative face-landmark SVG, evoking the MediaPipe mesh
 * the app uses for passive biometrics. Purely decorative: no real readings.
 */

// ~55 key landmark positions in a 300 × 380 coordinate space
const PTS = [
  // Face contour – left (0–9)
  [150, 18 ], // 0  crown
  [105, 38 ], // 1
  [68,  72 ], // 2
  [44, 120 ], // 3
  [38, 172 ], // 4
  [46, 224 ], // 5
  [68, 268 ], // 6
  [100,308 ], // 7
  [136,336 ], // 8
  [150,344 ], // 9  chin
  // Face contour – right (10–18)
  [164,336 ], // 10
  [200,308 ], // 11
  [232,268 ], // 12
  [254,224 ], // 13
  [262,172 ], // 14
  [256,120 ], // 15
  [232, 72 ], // 16
  [195, 38 ], // 17
  // Left eyebrow (18–22)
  [76, 104 ], // 18
  [100, 92 ], // 19
  [126, 88 ], // 20
  [148,100 ], // 21
  // Right eyebrow (22–25)
  [152,100 ], // 22
  [174, 88 ], // 23
  [200, 92 ], // 24
  [224,104 ], // 25
  // Left eye (26–31)
  [80, 134 ], // 26
  [102,122 ], // 27
  [128,118 ], // 28
  [150,132 ], // 29
  [128,149 ], // 30
  [102,151 ], // 31
  // Right eye (32–37)
  [220,134 ], // 32
  [198,122 ], // 33
  [172,118 ], // 34
  [150,132 ], // 35  shares with 29 visually; keep separate
  [172,149 ], // 36
  [198,151 ], // 37
  // Nose (38–45)
  [150,148 ], // 38  bridge top
  [150,172 ], // 39  mid bridge
  [150,196 ], // 40  tip
  [134,210 ], // 41
  [150,218 ], // 42
  [166,210 ], // 43
  [120,212 ], // 44  left nostril
  [180,212 ], // 45  right nostril
  // Mouth (46–55)
  [108,252 ], // 46  left corner
  [130,240 ], // 47
  [150,236 ], // 48  upper centre
  [170,240 ], // 49
  [192,252 ], // 50  right corner
  [170,268 ], // 51
  [150,276 ], // 52  lower centre
  [130,268 ], // 53
  // Cheek anchor points
  [60, 192 ], // 54  left cheek
  [240,192 ], // 55  right cheek
]

// Connections between landmark indices
const EDGES = [
  // Face contour
  [0,1],[1,2],[2,3],[3,4],[4,5],[5,6],[6,7],[7,8],[8,9],
  [9,10],[10,11],[11,12],[12,13],[13,14],[14,15],[15,16],[16,17],[17,0],
  // Left eyebrow
  [18,19],[19,20],[20,21],
  // Right eyebrow
  [22,23],[23,24],[24,25],
  // Left eye
  [26,27],[27,28],[28,29],[29,30],[30,31],[31,26],
  // Right eye
  [32,33],[33,34],[34,35],[35,36],[36,37],[37,32],
  // Nose bridge
  [38,39],[39,40],
  // Nose tip
  [40,41],[41,42],[42,43],[43,40],
  // Nostrils
  [44,41],[45,43],
  // Mouth outer
  [46,47],[47,48],[48,49],[49,50],[50,51],[51,52],[52,53],[53,46],
  // Mouth corners
  [46,50],
  // Structural – brow to eye
  [18,26],[21,29],[22,35],[25,32],
  // Structural – nose to eyes
  [38,28],[38,34],
  // Structural – cheek to contour
  [54,4],[54,5],[54,6],[55,13],[55,14],[55,12],
  // Structural – cheek to eyes
  [54,31],[55,37],
  // Structural – nose to mouth
  [44,46],[45,50],[40,48],
]

// Indices of "hero" points rendered larger with a highlight ring
const HERO_PTS = [28, 34, 40, 42, 48, 52, 9, 0]

export default function FaceMesh() {
  return (
    <div className="relative select-none">
      {/* Corner scan brackets */}
      {['top-0 left-0 border-t-2 border-l-2', 'top-0 right-0 border-t-2 border-r-2',
        'bottom-0 left-0 border-b-2 border-l-2', 'bottom-0 right-0 border-b-2 border-r-2']
        .map((cls, i) => (
          <div key={i} className={`absolute w-6 h-6 ${cls} border-accent/55`} />
        ))}

      <svg
        viewBox="0 0 300 380"
        width="340"
        height="430"
        aria-hidden="true"
        className="max-w-full"
      >
        <defs>
          {/* Vertical gradient for the scan line */}
          <linearGradient id="scanGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%"   stopColor="var(--color-accent)" stopOpacity="0"   />
            <stop offset="35%"  stopColor="var(--color-accent)" stopOpacity="0.7" />
            <stop offset="65%"  stopColor="var(--color-accent)" stopOpacity="0.7" />
            <stop offset="100%" stopColor="var(--color-accent)" stopOpacity="0"   />
          </linearGradient>

          {/* Soft glow filter for landmark dots */}
          <filter id="dotGlow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="1.5" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Face area clip for scan line */}
          <clipPath id="faceClip">
            <ellipse cx="150" cy="190" rx="130" ry="175" />
          </clipPath>

          {/* Subtle face fill */}
          <radialGradient id="faceAura" cx="50%" cy="46%" r="48%">
            <stop offset="0%"   stopColor="var(--color-accent)" stopOpacity="0.05" />
            <stop offset="100%" stopColor="var(--color-accent)" stopOpacity="0"    />
          </radialGradient>
        </defs>

        {/* Face area background glow */}
        <ellipse cx="150" cy="190" rx="130" ry="175" fill="url(#faceAura)" />

        {/* Mesh edges */}
        {EDGES.map(([a, b], i) => (
          <line
            key={i}
            x1={PTS[a][0]} y1={PTS[a][1]}
            x2={PTS[b][0]} y2={PTS[b][1]}
            stroke="var(--color-accent)"
            strokeWidth="0.45"
            strokeOpacity="0.22"
          />
        ))}

        {/* Landmark dots */}
        {PTS.map(([x, y], i) => (
          <circle
            key={i}
            cx={x} cy={y}
            r={HERO_PTS.includes(i) ? 2.5 : 1.6}
            fill="var(--color-accent)"
            opacity={HERO_PTS.includes(i) ? 0.85 : 0.45}
            filter="url(#dotGlow)"
          />
        ))}

        {/* Highlight rings on key points */}
        {HERO_PTS.map(i => (
          <circle
            key={`ring-${i}`}
            cx={PTS[i][0]} cy={PTS[i][1]}
            r="5"
            fill="none"
            stroke="var(--color-accent)"
            strokeWidth="0.7"
            opacity="0.5"
          />
        ))}

        {/* Animated scan line (SVG-native, frame-perfect) */}
        <g clipPath="url(#faceClip)">
          <rect x="20" y="-14" width="260" height="14" fill="url(#scanGrad)">
            <animateTransform
              attributeName="transform"
              type="translate"
              values="0,0; 0,400"
              dur="3.6s"
              repeatCount="indefinite"
              calcMode="linear"
            />
            <animate
              attributeName="opacity"
              values="0;1;1;0"
              keyTimes="0;0.06;0.92;1"
              dur="3.6s"
              repeatCount="indefinite"
            />
          </rect>
        </g>

      </svg>
    </div>
  )
}
