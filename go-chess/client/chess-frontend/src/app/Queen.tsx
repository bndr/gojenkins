export const Queen = ({fill, stroke}: {fill: string, stroke: string}) => (
    <svg xmlns="http://www.w3.org/2000/svg" width={45} height={45}>
      <g
        style={{
          fill: fill,
          stroke: stroke,
          strokeWidth: 1.5,
          strokeLinejoin: "round",
        }}
      >
        <path d="M9 26c8.5-1.5 21-1.5 27 0l2.5-12.5L31 25l-.3-14.1-5.2 13.6-3-14.5-3 14.5-5.2-13.6L14 25 6.5 13.5 9 26z" />
        <path d="M9 26c0 2 1.5 2 2.5 4 1 1.5 1 1 .5 3.5-1.5 1-1 2.5-1 2.5-1.5 1.5 0 2.5 0 2.5 6.5 1 16.5 1 23 0 0 0 1.5-1 0-2.5 0 0 .5-1.5-1-2.5-.5-2.5-.5-2 .5-3.5 1-2 2.5-2 2.5-4-8.5-1.5-18.5-1.5-27 0z" />
        <path
          d="M11.5 30c3.5-1 18.5-1 22 0M12 33.5c6-1 15-1 21 0"
          style={{
            fill: "none",
          }}
        />
        <circle cx={6} cy={12} r={2} />
        <circle cx={14} cy={9} r={2} />
        <circle cx={22.5} cy={8} r={2} />
        <circle cx={31} cy={9} r={2} />
        <circle cx={39} cy={12} r={2} />
      </g>
    </svg>
  )
  