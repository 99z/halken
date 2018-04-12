# halken
>Video games are meant to be just one thing. Fun. Fun for everyone.

*Satoru Iwata*

Halken is a Game Boy emulator written in Go being developed during my time at the [Recurse Center](https://recurse.com).
The name is one used by HAL Laboratories for a time. HAL was the first company where Satoru Iwata was a video game programmer.

## TODO

1. ~Implement CPU opcodes~
    * ~Non-CB opcodes implemented~
     * ~Write dispatch loop~
     * Use blargg's test output to fix instructions
      * 01 - special
      * 02 - interrupts
      * ~03 - op sp,hl~
      
      ![pass 3](https://my.mixtape.moe/gpwxlx.png)
      * ~04 - op r,imm~
      
      ![pass 4](https://my.mixtape.moe/glzofz.png)
      * ~05 - op rp~
      
      ![pass 5](https://my.mixtape.moe/rulxnw.png)
      * ~06 - ld r,r~
      
      ![pass 6](https://my.mixtape.moe/mfdkmk.png)
      * 07 - jr,jp,call,ret,rst
      * 08 - misc instrs
      * ~09 - op r,r~
      
      ![pass 9](https://my.mixtape.moe/jkitna.png)
      * ~10 - bit ops~
      
      ![pass 10](https://my.mixtape.moe/ysxqrh.png)
      * ~11 - op a,(hl)~
      
      ![pass 11](https://my.mixtape.moe/jiyqiu.png)
2. ~Implement memory~
3. ~Test GB bootstrap ROM~
4. ~Draw tiles~

   ![tile display](https://my.mixtape.moe/adxhwd.png)
5. ~Draw background~
6. ~Graphics loop~
