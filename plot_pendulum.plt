set terminal qt 0 persist
set title "Pendulum angle"
set xlabel "Time"
set ylabel "Angle"
set grid
plot "angle.txt" using 1:2 with lines title ""
pause -1

#set terminal qt 1 persist
#set title "Angular Speed"
#set xlabel "Time"
#set ylabel "Speed"
#set grid
#plot "speed.txt" using 1:2 with lines title ""
#pause -1
