import { useBrowserTheme } from "@/hooks";

const AuthBackdrop = () => {
  const { theme: browserTheme } = useBrowserTheme();

  const bgStyle = (() => {
    let svg = "";
    let mask = "";
    if (browserTheme === "dark") {
      svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="32" height="32" fill="none" stroke="rgb(202 78 13 / 0.12)"><path d="M0 .5H31.5V32"/></svg>`;
      mask = "white";
    } else {
      svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="32" height="32" fill="none" stroke="rgb(249 115 22 / 0.08)"><path d="M0 .5H31.5V32"/></svg>`;
      mask = "black";
    }

    return {
      backgroundImage: `url('data:image/svg+xml;base64,${btoa(svg)}')`,
      maskImage: `linear-gradient(to bottom right, transparent, ${mask}, transparent)`,
    } satisfies React.CSSProperties;
  })();

  return <div aria-hidden className="pointer-events-none fixed min-h-screen w-full" style={bgStyle}></div>;
};

export default AuthBackdrop;
