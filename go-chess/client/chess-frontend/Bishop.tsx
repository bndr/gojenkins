export const Bishop = ({fill, stroke}: {fill: string, stroke: string}) => (
    <svg xmlns="http://www.w3.org/2000/svg" width={45} height={45}>
      <g
        style={{
          opacity: 1,
          fill: "none",
          fillRule: "evenodd",
          fillOpacity: 1,
          stroke: stroke,
          strokeWidth: 1.5,
          strokeLinecap: "round",
          strokeLinejoin: "round",
          strokeMiterlimit: 4,
          strokeDasharray: "none",
          strokeOpacity: 1,
        }}
      >
        <g
          style={{
            fill: fill,
            stroke: stroke,
            strokeLinecap: "butt",
          }}
        >
          <path d="M9 36.6c3.39-.97 10.11.43 13.5-2 3.39 2.43 10.11 1.03 13.5 2 0 0 1.65.54 3 2-.68.97-1.65.99-3 .5-3.39-.97-10.11.46-13.5-1-3.39 1.46-10.11.03-13.5 1-1.35.49-2.32.47-3-.5 1.35-1.46 3-2 3-2z" />
          <path d="M15 32.6c2.5 2.5 12.5 2.5 15 0 .5-1.5 0-2 0-2 0-2.5-2.5-4-2.5-4 5.5-1.5 6-11.5-5-15.5-11 4-10.5 14-5 15.5 0 0-2.5 1.5-2.5 4 0 0-.5.5 0 2z" />
          <path d="M25 8.6a2.5 2.5 0 1 1-5 0 2.5 2.5 0 1 1 5 0z" />
        </g>
        <path
          d="M17.5 26h10M15 30h15m-7.5-14.5v5M20 18h5"
          style={{
            fill: "none",
            stroke: stroke,
            strokeLinejoin: "miter",
          }}
          transform="translate(0 .6)"
        />
      </g>
    </svg>
  )
  