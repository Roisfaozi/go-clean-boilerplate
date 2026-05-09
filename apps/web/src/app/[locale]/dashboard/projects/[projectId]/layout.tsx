import GoBack from "~/components/shared/go-back";

export default function SingleProjectLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <GoBack />
      {children}
    </>
  );
}
