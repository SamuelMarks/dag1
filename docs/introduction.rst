.. _introduction:

Introduction
============

What is DAG1?
---------------

DAG1 is an open-source software component intended for developers who want to 
build peer-to-peer (p2p) applications, mobile or other, without having to 
implement their own p2p networking layer from scratch. Under the hood, it 
enables many computers to behave as one; a technique known as state machine 
replication. 

DAG1 is designed to easily plug into applications written in any programming 
language. Developers can focus on building the application logic and simply 
integrate with DAG1 to handle the replication aspect. Basically, DAG1 will 
connect to other DAG1 nodes and guarantee that everyone processes the same 
commands in the same order. To do this, it uses p2p networking and a Byzantine 
Fault Tolerant (BFT) consensus algorithm.

.. image:: assets/dag1_network.png
   :height: 453px
   :width: 640px
   :align: left

DAG1 is:

- **Asynchronous**: 
    Participants have the freedom to process commands at different times.
- **Leaderless**: 
    No participant plays a 'special' role.
- **Byzantine Fault-Tolerant**: 
    Supports one third of faulty nodes, including malicious behavior.
- **Final**: 
    DAG1's output can be used immediately, no need for block confirmations, 
    etc.