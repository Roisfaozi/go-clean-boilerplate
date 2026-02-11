export function RenderIMGEl({
  title,
  logo,
  locale,
  image,
}: {
  title?: string;
  logo?: string;
  locale?: string;
  image?: string;
}) {
  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        width: "100%",
        height: "100%",
        backgroundColor: "white",
        backgroundImage:
          "linear-gradient(to bottom right, #E0E7FF 25%, #ffffff 50%, #CFFAFE 75%)",
      }}
    >
      {title && (
        <div
          style={{
            display: "flex",
            fontSize: 60,
            fontStyle: "normal",
            color: "black",
            marginTop: 30,
            lineHeight: 1.8,
            whiteSpace: "pre-wrap",
          }}
        >
          <b>{title}</b>
        </div>
      )}
      {locale && <div style={{ display: "flex", fontSize: 30 }}>{locale}</div>}
    </div>
  );
}
