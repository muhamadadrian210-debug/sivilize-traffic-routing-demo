import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Sivilize Traffic Routing Demo",
  description: "Sistem routing trafik berbasis aturan - Sivilize Corp",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body className="bg-slate-950 text-slate-100 min-h-screen antialiased">
        {children}
      </body>
    </html>
  );
}
