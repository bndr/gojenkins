"use client"
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { useState } from 'react'

const inter = Inter({ subsets: ['latin'] })



export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  
  return (
    <html className='w-screen h-full' lang="en">
      <body className="w-screen h-full">{children}</body>
    </html>
  )
}
