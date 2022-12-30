# epromtool

Simple console utility for 28c265-programmer.

## Flags

__-p=\<name>__ - port name  
__-b=\<baud>__ - baud rate  
__-u__ - unlock chip before clear/write  
__-c__ - clear chip before write  
__-l__ - lock chip after clear/write  
__-r__ - read chip  
__-w=\<file>__ - write file to chip

## Flags processing order

1. Read chip if __-r__ flag is set
2. Unlock chip if __-u__ flag is set
3. Clear chip if __-c__ flag is set
4. Write file to chip if __-w__ flag is set
5. Lock chip if __-l__ flag is set