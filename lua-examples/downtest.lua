local clock = os.clock
function sleep(n)
   local t0 = clock()
   while clock() - t0 <= n do
   end
end

sleep(40)
return 1 + 2