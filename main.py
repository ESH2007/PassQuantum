# ALL THE LIBS USED IN PASSQUANTUM
import base64, os, secrets, time
from quantcrypt.cipher import Krypton


# Encrypting function
def kryptonEncryption(text):
   # Krypton requires the symmetric secret key to be 64 bytes long.
   secret_key = secrets.token_bytes(64)
   krypton = Krypton(secret_key)

   # Encrypt the plaintext and generate the verification data packet.
   plaintext = text
   krypton.begin_encryption()
   ciphertext = krypton.encrypt(plaintext)
   verif_dp = krypton.finish_encryption()
   return ciphertext, verif_dp, secret_key


# Decrypting function
def kryptonDecryption(verif, ciphertext, secret_key):
   krypton = Krypton(secret_key)

   krypton.begin_decryption(verif)
   plaintext_copy = krypton.decrypt(ciphertext)
   krypton.finish_decryption()
   return str(plaintext_copy)


if __name__ == "__main__":
   while True:
      os.system('cls' if os.name == 'nt' else 'clear')  # Clears the screen in each iteration
      pass_list = []
      print("#################################################################################")
      print("|           Welcome To PassQuantum -The World's most save password manager-     |")
      print("#################################################################################")
      print("[If you want to add a new password, press \'Enter\']")
      print("[if you want to see you passwords type \"see passwords\", \"See Passwords\", \"seePass\" or \"2\"]")
      print("[If you want to exit the program, type \"exit\", \"e\", or \"3\"]")
      res = input("_> ")

      # Reads the passwords file and decrypt them so the user can see them
      if res == "see passwords" or res == "See Passwords" or res == "seePass" or res == "2":
         # Open the passwords file
         f = open("passwords.txt", "r")
         # splits the individual password information
         passwords_file = f.read().split(', \n')

         # This checks if the passwords file has any passwords
         if passwords_file[0] == '' or passwords_file[0] == "\n":
            print("There are no passwords in your password file yet")
            input("")
         else:
            print("Press \'Enter\' to return to the main menu")
            time.sleep(2)
            print("Decrypting your passwords...")
            time.sleep(2)
            print("Solving Schrödinger's Equation...")
            time.sleep(2)
            print("Solving Dirac's Equation...")
            time.sleep(2)
            print("Your passwords are decrypted! Below are your passwords: ")
            time.sleep(2)
            for password in passwords_file:
               if password != "": # This is to avoid error with the last line of the passwords file
                  # Splits the encrypted password, verification key and secret key
                  password_info = password.split(', ')
                  # This is to avoid that the password don't have any info. required to use kryptonDecryption()
                  if len(password_info) < 3:
                     print("ERROR: a password in your file don't have all the information required to decrypt it")
                     input("")
                  else:
                     # Decrypt the base64 characters back into bytes
                     encrypted_password = base64.b64decode(password_info[0][2:-1])
                     verif_key = base64.b64decode(password_info[1][2:-1])
                     priv_key = base64.b64decode(password_info[2][2:-1])
                     pass_decryp = kryptonDecryption(verif_key, encrypted_password, priv_key)
                     print(pass_decryp[2:-1]) # And then print each decrypted password with a 1 sec of delay
                     time.sleep(1)
            input("")
      elif res == "e" or res == "exit" or res == "3":
         break
      else:
         print("Press \'Enter\' to return to the main menu")
         res = input("Type a new password: ")

         time.sleep(2)
         print("Creating a random private key...")
         time.sleep(2)
         print("Putting Schrödinger's Cat inside the box...")

         # Encrypt and save the password in a file
         pass_encryp, verif_key, secret_key = kryptonEncryption(res)
         time.sleep(2)
         print(f"The private key is successfully created!\nYour private key is: {secret_key[:5]} \nYou better keep that safe!!")
         time.sleep(2)
         print("Encrypting your password...")
         time.sleep(2)
         print("Quantum interconnecting two photons...")
         time.sleep(2)
         print(f"Your password is encrypted! \nYour encrypted password is: {pass_encryp}")
         pass_list.append([pass_encryp, verif_key, secret_key])
         f = open("passwords.txt", 'a')
         f.write(
            f"[{base64.b64encode(pass_encryp)}, {base64.b64encode(verif_key)}, {base64.b64encode(secret_key)}], \n")
         f.close()
         input("")