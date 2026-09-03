--  <vc-preamble>
with Ada.Numerics.Big_Numbers.Big_Integers;
use  Ada.Numerics.Big_Numbers.Big_Integers;

package Np_Prod_Spec with SPARK_Mode is

   --  Bounds are chosen so that the product of a whole array stays inside
   --  Integer: Max_Value ** (Max_Index + 1) <= Integer'Last.
   Max_Index   : constant := 13;
   Max_Value   : constant := 4;
   Max_Product : constant := Max_Value ** (Max_Index + 1);

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;
   subtype Product_Type is Integer range -Max_Product .. Max_Product;

   --  A bound may sit one past the last index, so that an empty slice
   --  Start .. Finish - 1 can be written for Start = Finish = A'Last + 1.
   subtype Bound_Type is Natural range 0 .. Max_Index + 1;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Mathematical product of every element of A, empty product = 1.
   --  Computed over Big_Integer so that the specification itself is exact
   --  and free of overflow, exactly as Dafny's unbounded ints are.
   function Prod_All (A : Int_Array) return Big_Integer is
     (if A'Length = 0 then To_Big_Integer (1)
      else To_Big_Integer (A (A'First)) * Prod_All (A (A'First + 1 .. A'Last)))
   with Subprogram_Variant => (Decreases => A'Length);
--  </vc-preamble>

--  <vc-spec>
   procedure Prod (A : Int_Array; Result : out Product_Type) with
     Post => To_Big_Integer (Result) = Prod_All (A);

   procedure Prod_Array
     (A      : Int_Array;
      Start  : Bound_Type;
      Finish : Bound_Type;
      Result : out Product_Type)
   with
     Pre  => Start >= A'First and then Start <= Finish
             and then Finish <= A'Last + 1,
     Post => To_Big_Integer (Result) = Prod_All (A (Start .. Finish - 1));

end Np_Prod_Spec;

with Ada.Numerics.Big_Numbers.Big_Integers;
use  Ada.Numerics.Big_Numbers.Big_Integers;

package body Np_Prod_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Prod (A : Int_Array; Result : out Product_Type) is
   begin
      pragma Assume (False);
   end Prod;

   procedure Prod_Array
     (A      : Int_Array;
      Start  : Bound_Type;
      Finish : Bound_Type;
      Result : out Product_Type) is
   begin
      pragma Assume (False);
   end Prod_Array;
--  </vc-code>

--  <vc-postamble>
end Np_Prod_Spec;
--  </vc-postamble>
