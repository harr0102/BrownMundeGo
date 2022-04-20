# BrownMundeGo
Made by
Shuja Hussain (shhu@itu.dk) and 
Harry Singh (hars@itu.dk)

<h3> Description </h3>
<p> Soon to be written. </p>
<br/>

<h2> How to use Man-in-the-Middle attack</h2>
<p>To use our program in an attack scenario you would have to: </p>
<ol>
  <li>Find the MAC address of the OBD2 Dongle device in the car.</li>
  <p>This is done by running file <b>discoverMAC.go</b></p>
  <code>sudo go run ./discoverMAC.go</code>
  <p>The program will look for the bluetooth device. If found it will inform you and create a file: <b>macSpoof.sh</b>. This bash file includes commands that will change the Raspberry PI bluetooth MAC address. <b>To reset the MAC address: restart the Raspberry PI.</b></p>
</ol>

<br />
When running main.go a server opens and a website is hosted on http://YOUR-SERVER-LOCAL-ADDRESS:8080. This website provides a user-friendly webpage for executing commands towards the OBD2-dongle.
