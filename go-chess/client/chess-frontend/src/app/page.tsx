"use client"

import Image from 'next/image'
import React, { ReactNode, useEffect, useState } from 'react'
import {BlackPawn, Pawn, WhitePawn} from  './Pawn'
import King from './King'
import { Queen } from './Queen'
import { Bishop } from '../../Bishop'
import { Knight } from './Knight'
import { Rook } from './Rook'

const CHESS_BOARD_SIZE = 8


export default function Home() {
  const [gameBoardState, setGameBoardState] = useState<string[]>([]);
  let gameBoard: string[] = []
  for(let i = 0; i < CHESS_BOARD_SIZE; i++ ){
    for(let j = 0; j < CHESS_BOARD_SIZE; j++){
      if((i + j) % 2 == 0){
        gameBoard.push("bg-white")
      }
      else{
        gameBoard.push("bg-sky-800")
      }
    }
  }

  useEffect(() => {
    setGameBoardState(gameBoard)
  }, [gameBoard])
  
  return (
    <main className='w-full flex items-center place-content-center h-screen'>
      <div className="w-1/2 h-auto p-6 bg-red-900 flex flex-col">
          <div className='bg-gray-900 flex h-20'>
            <div className='text-center m-auto'>
              <p>Aidans Wonderful World of Chess</p>
            </div>
          </div>
          {GameBoard(gameBoardState)}
      </div>
    </main>
  )
}

function GamePiece({Piece}: {Piece: ReactNode}){
  return ( 
  <div className='m-auto text-center flex place-content-center'>
  {Piece}
  </div>
  )
}



const SetupPieces = () => {
  const pieces: React.JSX.Element[] = []
  SetupPawns(pieces)

  // Black King
  pieces[3] = <GamePiece Piece={<King fill="black" stroke="white"></King>}></GamePiece>

  // White King
  pieces[59] = <GamePiece Piece={<King fill="white" stroke="black"></King>}></GamePiece>

  // Black Queen
  pieces[4] = <GamePiece Piece={<Queen fill="black" stroke="white"></Queen>}></GamePiece>

  // White Queen (.)(.)
  pieces[60] = <GamePiece Piece={<Queen fill="white" stroke ="black"></Queen>}></GamePiece>

  // Black Bishop 1
  pieces[2] = <GamePiece Piece={<Bishop fill="black" stroke="white"></Bishop>}></GamePiece>
  pieces[5] = <GamePiece Piece={<Bishop fill="black" stroke="white"></Bishop>}></GamePiece>

  pieces[58] = <GamePiece Piece={<Bishop fill="white" stroke="black"></Bishop>}></GamePiece>
  pieces[61] = <GamePiece Piece={<Bishop fill="white" stroke="black"></Bishop>}></GamePiece>

  //KNIGHTS BLACK DICK ENERGY
  pieces[6] = <GamePiece Piece={<Knight fill="black" stroke="white"></Knight>}></GamePiece>
  pieces[1] = <GamePiece Piece={<Knight fill="black" stroke="white"></Knight>}></GamePiece>

  pieces[57] = <GamePiece Piece={<Knight fill="white" stroke="black"></Knight>}></GamePiece>
  pieces[62] = <GamePiece Piece={<Knight fill="white" stroke="black"></Knight>}></GamePiece>
  // DONGLEY CASTLE ENERGY
  pieces[0] = <GamePiece Piece={<Rook fill="black" stroke="white"></Rook>}></GamePiece>
  pieces[7] = <GamePiece Piece={<Rook fill="black" stroke="white"></Rook>}></GamePiece> 

  pieces[56] = <GamePiece Piece={<Rook fill="white" stroke="black"></Rook>}></GamePiece>
  pieces[63] = <GamePiece Piece={<Rook fill="white" stroke="black"></Rook>}></GamePiece> 
  return pieces

}

const SetupPawns = (pieces: React.JSX.Element[]) => {
  const BLACK_PAWN_START = 8
  const BLACK_PAWN_END = 16

  for(let i = BLACK_PAWN_START; i < BLACK_PAWN_END; i++){
    pieces[i] = <GamePiece Piece={<BlackPawn></BlackPawn>}></GamePiece>
  }

  const WHITE_PAWN_START = 48
  const WHITE_PAWN_END = 56 

  for(let i = WHITE_PAWN_START; i < WHITE_PAWN_END; i++){
    pieces[i] = <GamePiece Piece={<WhitePawn></WhitePawn>}></GamePiece>
  }

}


const GameBoard = (gameBoardState: any[]) => {
  const pieces = SetupPieces()
  

  return(
    <div className="grid grid-cols-8 place-content-center gap-0 w-full">
      {gameBoardState.map((val, i) => 
      <div key={i} className={`${val} ${val === "bg-white" ? `hover:bg-slate-200` : `hover:bg-gray-800`} aspect-square w-full flex place-content-center`}>
        {pieces[i]}
      </div>)
      }

    </div>
  )
}
