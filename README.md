# BrownMundeGo
Made by
Shuja Hussain (shhu@itu.dk) and 
Harry Singh (hars@itu.dk)

<h2> How to use Man-in-the-Middle attack</h2>
<p>To use our program in an attack scenario you would have to: </p>
<ol>
  <li>Find the MAC address of the OBD2 Dongle device in the car.</li>
  <code>sudo go run ./discoverMAC.go</code>
  <p>The program will look for the bluetooth device. If found it will inform you and create a file: <b>macSpoof.sh</b>. This bash file includes commands that will change the Raspberry PI bluetooth MAC address and to run the .mitmAttack.go. <b>To reset the MAC address: restart the Raspberry PI.</b></p>
  
  <li>Run the Man-in-the-Middle attack program without discoverMAC</li>
  <b>Run with server</b><br/>
  <code>sudo go run ./mitmAttack.go XX:XX:XX:XX:XX:XX</code>
  <p>XX being the MAC address.
</ol>


<h2> How to attack OBD2-dongle</h2>
<p>If you wish to execute commands towards the OBD2-dongle. This can be used to read fuel level, voltage, etc.</p>
<code>sudo go run ./attackDongle.go</code>
<p>When running attackDongle.go: a server opens and a website is hosted on http://YOUR-SERVER-LOCAL-ADDRESS:8080. This website provides a user-friendly webpage for executing DongleScope attack.</p>
