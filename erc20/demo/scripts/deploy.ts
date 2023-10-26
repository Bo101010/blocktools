import { ethers } from "hardhat";

async function main() {
  const supply = ethers.parseEther("100000000");

  const demo = await ethers.deployContract("DemoToken", [supply]);
  const DEMO = await demo.waitForDeployment();
  console.log(await DEMO.getAddress());
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});

// https://goerli.etherscan.io/address/0x5c486db7559adAC22516Ae7676750f5105A1F3d1
