#!/usr/bin/python3

import logging
import time
logging.basicConfig(format='%(asctime)s %(levelname)s %(message)s', level="DEBUG")
n = 200
i = 0

print("Starting Loop")
logging.debug("Starting Loop")
while i <= n:
#while True:
   print("Counter is: {} of {}".format(i, n))
   logging.debug("Counter is: {} of {}".format(i, n))
   i = i+1
   time.sleep(20)
print("completed the loop terminating")
logging.debug("completed the loop terminating")
